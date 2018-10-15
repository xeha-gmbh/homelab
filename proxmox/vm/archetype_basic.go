package vm

import (
	"fmt"
	"github.com/imulab/homelab/proxmox/common"
	"github.com/lithammer/dedent"
	"github.com/spf13/pflag"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	basicArchFlagNode 			  = "node"
	basicArchFlagVmId             = "id"
	basicArchFlagName             = "name"
	basicArchFlagIsoStorage       = "iso-storage"
	basicArchFlagIsoImage         = "iso-image"
	basicArchFlagDriveStorage     = "drive-storage"
	basicArchFlagDriveSize        = "drive-size"
	basicArchFlagCore             = "core"
	basicArchFlagMemory           = "memory"
	basicArchFlagNetworkInterface = "iface"
	basicArchFlagStart			  = "start"

	basicArchDefaultNode = "pve"
	basicArchDefaultIsoStorage = "local"
	basicArchDefaultDriveSize = 64
	basicArchDefaultCore = 2
	basicArchDefaultMemory = 2048
	basicArchDefaultNetworkIFace = "vmbr0"
	basicArchDefaultStart = false

	noDefault = ""
)

func init() {
	ArchetypeRepository().SubmitArchetype(&basicArchetype{})
}

type basicArchetype struct {
	node 			string
	vmId			string
	vmName 			string
	isoStorage		string
	isoImage 		string
	driveStorage	string
	driveSize		int
	cpuCores		int
	memory			int
	networkIFace 	string
	start 			bool
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

func (b *basicArchetype) BindFlags(flagSet *pflag.FlagSet) {
	flagSet.StringVar(
		&b.node, basicArchFlagNode, basicArchDefaultNode,
		"The node which VM will be created on.",
	)
	flagSet.StringVar(
		&b.vmId, basicArchFlagVmId, noDefault,
		"The ID number of the new VM. Must be unique. Required.",
		)
	flagSet.StringVar(
		&b.vmName, basicArchFlagName, noDefault,
		"The name of the new VM. Required.",
		)
	flagSet.StringVar(
		&b.isoStorage, basicArchFlagIsoStorage, basicArchDefaultIsoStorage,
		"The storage device name for the ISO installation media.",
		)
	flagSet.StringVar(
		&b.isoImage, basicArchFlagIsoImage, noDefault,
		"File name for the ISO installation media. Required.",
	)
	flagSet.StringVar(
		&b.driveStorage, basicArchFlagDriveStorage, noDefault,
		"The storage device name for the hard drive. Required.",
	)
	flagSet.IntVar(
		&b.driveSize, basicArchFlagDriveSize, basicArchDefaultDriveSize,
		"The size in GB of the hard drive.",
		)
	flagSet.IntVar(
		&b.cpuCores, basicArchFlagCore, basicArchDefaultCore,
		"Number of of virtual CPU cores.",
	)
	flagSet.IntVar(
		&b.memory, basicArchFlagMemory, basicArchDefaultMemory,
		"Amount of virtual memory in MB",
	)
	flagSet.StringVar(
		&b.networkIFace, basicArchFlagNetworkInterface, basicArchDefaultNetworkIFace,
		"Host interface to bridge the network to.",
	)
	flagSet.BoolVar(
		&b.start, basicArchFlagStart, basicArchDefaultStart,
		"Starts VM after successful creation.")
}

// Post to Proxmox API to create a VM. If '--start' is requested, it will attempt to start the VM.
func (b *basicArchetype) CreateVM() error {
	if err := b.doCreateVM(); err != nil {
		return err
	} else {
		if b.start {
			return b.startVM()
		}
		return nil
	}
}

func (b *basicArchetype) doCreateVM() error {
	var (
		err 		error
		subject		*common.ProxmoxSubject
		req 		*http.Request
		resp 		*http.Response
	)

	if subject, err = common.ReadSubjectFromCache(); err != nil {
		return common.GenericError(fmt.Errorf("failed to read ticket cache: %s\n", err.Error()))
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
		return common.GenericError(err)
	} else if req, err = common.WithHttpCredentials(req); err != nil {
		return common.ProxmoxError(fmt.Errorf("unable to locate session: %s", err.Error()))
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := common.HttpClient()
	if resp, err = client.Do(req); err != nil {
		return common.ProxmoxError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return common.ProxmoxError(fmt.Errorf("failed to create VM: %s\n", resp.Status))
	}

	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return common.GenericError(err)
	} else {
		fmt.Fprintln(os.Stdout, string(body))
	}

	return nil
}

func (b *basicArchetype) startVM() error {
	var (
		err 		error
		subject		*common.ProxmoxSubject
		req 		*http.Request
		resp 		*http.Response
	)

	if subject, err = common.ReadSubjectFromCache(); err != nil {
		return common.GenericError(fmt.Errorf("failed to read ticket cache: %s\n", err.Error()))
	}

	if req, err = http.NewRequest(http.MethodPost, qemuStartUrl(subject.ApiServer, b.node, b.vmId), nil); err != nil {
		return common.GenericError(err)
	} else if req, err = common.WithHttpCredentials(req); err != nil {
		return common.ProxmoxError(err)
	}

	client := common.HttpClient()
	if resp, err = client.Do(req); err != nil {
		return common.ProxmoxError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return common.ProxmoxError(fmt.Errorf("failed to start VM: %s\n", resp.Status))
	}

	fmt.Fprintf(os.Stdout, "VM %s started.\n", b.vmId)
	return nil
}

func qemuUrl(base, node string) string {
	return fmt.Sprintf("%s/api2/json/nodes/%s/qemu", base, node)
}

func qemuStartUrl(base, node, vmId string) string {
	return fmt.Sprintf("%s/api2/json/nodes/%s/qemu/%s/status/start", base, node, vmId)
}