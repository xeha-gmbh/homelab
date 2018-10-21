package vm

import (
	"fmt"
	"github.com/imulab/homelab/proxmox/common"
	"github.com/imulab/homelab/shared"
	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	basicArchFlagNode             = "node"
	basicArchFlagVmId             = "id"
	basicArchFlagName             = "name"
	basicArchFlagIsoStorage       = "iso-storage"
	basicArchFlagIsoImage         = "iso-image"
	basicArchFlagDriveStorage     = "drive-storage"
	basicArchFlagDriveSize        = "drive-size"
	basicArchFlagCore             = "core"
	basicArchFlagMemory           = "memory"
	basicArchFlagNetworkInterface = "iface"
	basicArchFlagStart            = "start"

	basicArchDefaultNode         = "pve"
	basicArchDefaultIsoStorage   = "local"
	basicArchDefaultDriveSize    = 64
	basicArchDefaultCore         = 2
	basicArchDefaultMemory       = 2048
	basicArchDefaultNetworkIFace = "vmbr0"
	basicArchDefaultStart        = false

	noDefault = ""
)

func init() {
	ArchetypeRepository().SubmitArchetype(&basicArchetype{})
}

type basicArchetype struct {
	shared.ExtraArgs
	_output      shared.MessagePrinter
	node         string
	vmId         string
	vmName       string
	isoStorage   string
	isoImage     string
	driveStorage string
	driveSize    int
	cpuCores     int
	memory       int
	networkIFace string
	start        bool
}

func (b *basicArchetype) Name() string {
	return "basic archetype"
}

func (b *basicArchetype) Use() string {
	return "basic"
}

func (b *basicArchetype) Short() string {
	return "basic archetype to create a vm."
}

func (b *basicArchetype) Long() string {
	return dedent.Dedent(`
		This archetype describes a VM with basic configuration. It has the following features:
			* Creates Linux VM with 2.6/3.X or later Kernal.
			* Creates VM with one CPU socket and configurable number of cores.
			* Creates one hard drive with SCSI format with VirtIO PCI driver.
			* Creates bridged network with VirtIO driver on the configurable interface.
			* Installs OS using ISO image mounted as CD-ROM.
			* NUMA support is turned on.
		
		This archetype supports most of my workstation needs.
	`)
}

func (b *basicArchetype) AllFlags() []string {
	return []string{
		basicArchFlagNode,
		basicArchFlagVmId,
		basicArchFlagName,
		basicArchFlagIsoStorage,
		basicArchFlagIsoImage,
		basicArchFlagDriveStorage,
		basicArchFlagDriveSize,
		basicArchFlagCore,
		basicArchFlagMemory,
		basicArchFlagNetworkInterface,
		basicArchFlagStart,
	}
}

func (b *basicArchetype) RequiredFlags() []string {
	return []string{
		basicArchFlagVmId,
		basicArchFlagName,
		basicArchFlagIsoImage,
		basicArchFlagDriveStorage,
	}
}

func (b *basicArchetype) BindFlags(cmd *cobra.Command) {
	b.InjectExtraArgs(cmd)
	b._output = shared.WithConfig(cmd, &b.ExtraArgs)

	cmd.Flags().StringVar(
		&b.node, basicArchFlagNode, basicArchDefaultNode,
		"The node which VM will be created on.",
	)
	cmd.Flags().StringVar(
		&b.vmId, basicArchFlagVmId, noDefault,
		"The ID number of the new VM. Must be unique. Required.",
	)
	cmd.Flags().StringVar(
		&b.vmName, basicArchFlagName, noDefault,
		"The name of the new VM. Required.",
	)
	cmd.Flags().StringVar(
		&b.isoStorage, basicArchFlagIsoStorage, basicArchDefaultIsoStorage,
		"The storage device name for the ISO installation media.",
	)
	cmd.Flags().StringVar(
		&b.isoImage, basicArchFlagIsoImage, noDefault,
		"File name for the ISO installation media. Required.",
	)
	cmd.Flags().StringVar(
		&b.driveStorage, basicArchFlagDriveStorage, noDefault,
		"The storage device name for the hard drive. Required.",
	)
	cmd.Flags().IntVar(
		&b.driveSize, basicArchFlagDriveSize, basicArchDefaultDriveSize,
		"The size in GB of the hard drive.",
	)
	cmd.Flags().IntVar(
		&b.cpuCores, basicArchFlagCore, basicArchDefaultCore,
		"Number of of virtual CPU cores.",
	)
	cmd.Flags().IntVar(
		&b.memory, basicArchFlagMemory, basicArchDefaultMemory,
		"Amount of virtual memory in MB",
	)
	cmd.Flags().StringVar(
		&b.networkIFace, basicArchFlagNetworkInterface, basicArchDefaultNetworkIFace,
		"Host interface to bridge the network to.",
	)
	cmd.Flags().BoolVar(
		&b.start, basicArchFlagStart, basicArchDefaultStart,
		"Starts VM after successful creation.")
}

// Post to Proxmox API to create a VM. If '--start' is requested, it will attempt to start the VM.
func (b *basicArchetype) CreateVM() error {
	if err := b.doCreateVM(); err != nil {
		b._output.Fatal(shared.ErrOp.ExitCode,
			"failed to create vm {{index .id}} on proxmox. Cause: {{index .cause}}",
			map[string]interface{}{
				"event": "vm_creation_failed",
				"id":    b.vmId,
				"cause": err.Error(),
			})
		return shared.ErrOp
	}

	if b.start {
		if err := b.startVM(); err != nil {
			b._output.Fatal(shared.ErrOp.ExitCode,
				"failed to start vm {{index .id}} on proxmox. Cause: {{index .cause}}",
				map[string]interface{}{
					"event": "vm_start_failed",
					"id":    b.vmId,
					"cause": err.Error(),
				})
			return err
		}
	}

	b._output.Info("vm {{index .id}} is created on proxmox.",
		map[string]interface{}{
			"event": "vm_creation_success",
			"id":    b.vmId,
		})
	return nil
}

func (b *basicArchetype) doCreateVM() error {
	var (
		err     error
		subject *common.ProxmoxSubject
		req     *http.Request
		resp    *http.Response
	)

	if subject, err = common.ReadSubjectFromCache(); err != nil {
		return fmt.Errorf("unable to read ticket: %s", err.Error())
	}

	form := url.Values{}
	form.Set("vmid", b.vmId)
	form.Set("name", b.vmName)
	form.Set("ide2", fmt.Sprintf("%s:iso/%s,media=cdrom", b.isoStorage, b.isoImage))
	form.Set("ostype", "l26")
	form.Set("scsihw", "virtio-scsi-pci")
	form.Set("scsi0", fmt.Sprintf("%s:%d", b.driveStorage, b.driveSize))
	form.Set("sockets", "1")
	form.Set("cores", fmt.Sprintf("%d", b.cpuCores))
	form.Set("numa", "1")
	form.Set("memory", fmt.Sprintf("%d", b.memory))
	form.Set("net0", fmt.Sprintf("virtio,bridge=vmbr0"))

	if req, err = http.NewRequest(http.MethodPost, qemuUrl(subject.ApiServer, b.node), strings.NewReader(form.Encode())); err != nil {
		return err
	} else if req, err = common.WithHttpCredentials(req); err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := common.HttpClient()
	if resp, err = client.Do(req); err != nil {
		return err
	}
	defer resp.Body.Close()

	b._output.Debug("create vm request status: {{index .code}}.",
		map[string]interface{}{
			"event":  "http_response",
			"code":   resp.StatusCode,
			"status": resp.Status,
		})

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("create vm request non-200 code: %d", resp.StatusCode)
	}

	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return shared.ErrParse
	} else {
		b._output.Debug("http response body:\n\n{{index .content}}\n",
			map[string]interface{}{
				"event":   "http_response",
				"content": string(body),
			})
	}

	return nil
}

func (b *basicArchetype) startVM() error {
	var (
		err     error
		subject *common.ProxmoxSubject
		req     *http.Request
		resp    *http.Response
	)

	if subject, err = common.ReadSubjectFromCache(); err != nil {
		return fmt.Errorf("unable to read ticket: %s", err.Error())
	}

	if req, err = http.NewRequest(http.MethodPost, qemuStartUrl(subject.ApiServer, b.node, b.vmId), nil); err != nil {
		return err
	} else if req, err = common.WithHttpCredentials(req); err != nil {
		return err
	}

	client := common.HttpClient()
	if resp, err = client.Do(req); err != nil {
		return err
	}
	defer resp.Body.Close()

	b._output.Debug("start vm request status: {{index .code}}.",
		map[string]interface{}{
			"event":  "http_response",
			"code":   resp.StatusCode,
			"status": resp.Status,
		})

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("start vm request non-200 code: %d", resp.StatusCode)
	}

	return nil
}

func qemuUrl(base, node string) string {
	return fmt.Sprintf("%s/api2/json/nodes/%s/qemu", base, node)
}

func qemuStartUrl(base, node, vmId string) string {
	return fmt.Sprintf("%s/api2/json/nodes/%s/qemu/%s/status/start", base, node, vmId)
}
