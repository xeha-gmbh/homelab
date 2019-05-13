package bootstrap

import (
	"errors"
	"fmt"
	"github.com/xeha-gmbh/homelab/shared"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// The proxmox provider
type proxmoxProvider struct {
	Api      string `yaml:"api"`
	Identity struct {
		Realm    string `yaml:"realm"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"identity"`
	DataStores []struct {
		Name string   `yaml:"name"`
		Tags []string `yaml:"tags"`
	} `yaml:"datastores"`
}

func (p *proxmoxProvider) Name() string {
	return proxmox
}

func (p *proxmoxProvider) CreateVM(vm *VM, images []*Image) error {
	var (
		err           error
		dlImagePath   string
		autoImagePath string
		image         *Image
	)

	if image, err = p.getImage(vm.Image.Name, images); err != nil {
		return err
	}

	output.Info("Ensuring image {{index .imageName}} exists. Necessary downloads may take a while.",
		map[string]interface{}{
			"event":     "pre_ensure_image",
			"imageName": image.Name,
		})
	if dlImagePath, err = p.ensureImage(vm, image); err != nil {
		return err
	}
	output.Info("Image {{index .imageName}} now exists at {{index .path}}",
		map[string]interface{}{
			"event":     "post_ensure_image",
			"imageName": image.Name,
			"path":      dlImagePath,
		})

	output.Info("Processing image {{index .path}}.",
		map[string]interface{}{
			"event": "pre_process_image",
			"path":  dlImagePath,
		})
	if autoImagePath, err = p.createAutoInstallImage(vm, image, dlImagePath); err != nil {
		return err
	}
	output.Info("Processed image. New image at {{index .path}}",
		map[string]interface{}{
			"event": "post_process_image",
			"path":  autoImagePath,
		})

	output.Info("Uploading image {{index .path}}.",
		map[string]interface{}{
			"event": "pre_upload_image",
			"path":  autoImagePath,
		})
	if err = p.uploadAutoInstallImage(vm, image, autoImagePath); err != nil {
		return err
	}
	output.Info("Image {{index .path}} uploaded.",
		map[string]interface{}{
			"event": "post_upload_image",
			"path":  autoImagePath,
		})

	output.Info("Creating VM {{index .id}}.",
		map[string]interface{}{
			"event": "pre_create_vm",
			"id":    vm.Id,
		})
	if err = p.createAndStartVM(vm, autoImagePath); err != nil {
		return err
	}
	output.Info("VM {{index .id}} created.",
		map[string]interface{}{
			"event": "post_create_vm",
			"id":    vm.Id,
		})

	return nil
}

func (p *proxmoxProvider) createAndStartVM(vm *VM, filePath string) error {
	var err error

	if err = p.ensureLoggedIn(vm); err != nil {
		return err
	} else {
		output.Info("User logged in.", map[string]interface{}{})
	}

	proxmoxVmCreateArgs := []string{
		"proxmox",
		"vm",
		"create",
		vm.Archetype,
		"--output-format", shared.OutputFormatJson,
		"--start",
	}
	switch vm.Archetype {
	case basicArchetype:
		params := vm.Params.(*proxmoxBasicArchetypeParams)
		proxmoxVmCreateArgs = append(proxmoxVmCreateArgs, []string{
			"--id", vm.Id,
			"--name", vm.Name,
			"--node", vm.Provider.Args["node"].(string),
			"--core", fmt.Sprintf("%d", params.Cpu),
			"--memory", fmt.Sprintf("%d", params.MemoryMB()),
			"--drive-size", fmt.Sprintf("%d", params.DriveGB()),
			"--drive-storage", params.Drive.Store,
			"--iso-image", filepath.Base(filePath),
			"--iso-storage", vm.Image.Store,
			"--iface", params.Network.Interface,
		}...)
	default:
		return fmt.Errorf("unknown archetype %s", vm.Archetype)
	}
	proxmoxVmCreate := exec.Command("homelab", proxmoxVmCreateArgs...)

	_, err = shared.HandleOutput(output)(proxmoxVmCreate.CombinedOutput())(func(data map[string]interface{}) (interface{}, error) {
		if len(data) > 0 {
			if strings.ToUpper(data["level"].(string)) == "ERROR" {
				return nil, errors.New(data["message"].(string))
			}
		}
		return nil, nil
	})

	return err
}

func (p *proxmoxProvider) uploadAutoInstallImage(vm *VM, image *Image, filePath string) error {
	var err error

	if err = p.ensureLoggedIn(vm); err != nil {
		return err
	} else {
		output.Info("User logged in.", map[string]interface{}{})
	}

	proxmoxUploadArgs := []string{
		"proxmox",
		"upload",
		"--node", vm.Provider.Args["node"].(string),
		"--file", filePath,
		"--format", image.Format,
		"--node", vm.Provider.Args["node"].(string),
		"--storage", vm.Image.Store,
		"--output-format", shared.OutputFormatJson,
	}
	if extraArgs.Debug {
		proxmoxUploadArgs = append(proxmoxUploadArgs, "--debug")
	}
	proxmoxUpload := exec.Command("homelab", proxmoxUploadArgs...)

	_, err = shared.HandleOutput(output)(proxmoxUpload.CombinedOutput())(func(data map[string]interface{}) (interface{}, error) {
		if len(data) > 0 {
			if strings.ToUpper(data["level"].(string)) == "ERROR" {
				return nil, errors.New(data["message"].(string))
			}
		}
		return nil, nil
	})

	return err
}

func (p *proxmoxProvider) ensureLoggedIn(vm *VM) error {
	var err error

	proxmoxLoginArgs := []string{
		"proxmox",
		"login",
		"--username", p.Identity.Username,
		"--password", p.Identity.Password,
		"--realm", p.Identity.Realm,
		"--api-server", p.Api,
		"--force",
		"--output-format", shared.OutputFormatJson,
	}
	if vm.Provider.Args["force-login"].(bool) {
		proxmoxLoginArgs = append(proxmoxLoginArgs, "--force")
	}
	if extraArgs.Debug {
		proxmoxLoginArgs = append(proxmoxLoginArgs, "--debug")
	}
	proxmoxLogin := exec.Command("homelab", proxmoxLoginArgs...)

	_, err = shared.HandleOutput(output)(proxmoxLogin.CombinedOutput())(func(data map[string]interface{}) (interface{}, error) {
		if len(data) > 0 {
			if strings.ToUpper(data["level"].(string)) == "ERROR" {
				return nil, errors.New(data["message"].(string))
			}
		}
		return nil, nil
	})

	return err
}

func (p *proxmoxProvider) createAutoInstallImage(vm *VM, image *Image, downloadedImagePath string) (string, error) {
	var (
		err        error
		outputPath = filepath.Join(
			tempDir,
			fmt.Sprintf("%s-%s.iso",
				strings.Replace(image.Flavor, string(filepath.Separator), "-", -1),
				vm.Id))
	)

	if !image.Auto {
		return downloadedImagePath, nil
	}

	isoAutoArgs := []string{
		"iso",
		"auto",
		"--flavor", image.Flavor,
		"--input-iso", downloadedImagePath,
		"--output-iso", outputPath,
		"--workspace", tempDir,
		"--output-format", shared.OutputFormatJson,
	}
	if image.UsbBoot {
		isoAutoArgs = append(isoAutoArgs, "--usb-boot")
	}
	if image.Reuse {
		isoAutoArgs = append(isoAutoArgs, "--reuse")
	}
	if extraArgs.Debug {
		isoAutoArgs = append(isoAutoArgs, "--debug")
	}
	switch vm.Archetype {
	case basicArchetype:
		params := vm.Params.(*proxmoxBasicArchetypeParams)
		isoAutoArgs = append(isoAutoArgs, []string{
			"--timezone", params.System.Timezone,
			"--username", params.System.Username,
			"--password", params.System.Password,
			"--hostname", params.System.Hostname,
			"--domain", params.System.Domain,
			"--ip-address", params.Network.Ip,
			"--net-mask", params.Network.Mask,
			"--gateway", params.Network.Gateway,
			"--name-servers", strings.Join(params.Network.Dns, " "),
		}...)
	default:
		return "", fmt.Errorf("unknown archetype %s", vm.Archetype)
	}
	isoAuto := exec.Command("homelab", isoAutoArgs...)

	result, err := shared.HandleOutput(output)(isoAuto.CombinedOutput())(func(data map[string]interface{}) (interface{}, error) {
		if len(data) > 0 {
			switch strings.ToLower(data["event"].(string)) {
			case "remaster-success":
				return data["outputPath"], nil
			default:
				if strings.ToUpper(data["level"].(string)) == "ERROR" {
					return nil, errors.New(data["message"].(string))
				} else {
					return nil, unknownReturnStatus
				}
			}
		}
		return "", unknownReturnStatus
	})

	if err != nil {
		return "", err
	}
	return result.(string), nil
}

func (p *proxmoxProvider) copy(source, dest string) error {
	var (
		in, out *os.File
		err     error
	)

	if in, err = os.Open(source); err != nil {
		return err
	}

	if out, err = os.Create(dest); err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	return err
}

func (p *proxmoxProvider) ensureImage(vm *VM, image *Image) (file string, err error) {
	isoGetArgs := []string{
		"iso",
		"get",
		"--flavor", image.Flavor,
		"--target-dir", tempDir,
		"--output-format", shared.OutputFormatJson,
		"--reuse",
	}
	isoGet := exec.Command("homelab", isoGetArgs...)

	result, err := shared.HandleOutput(output)(isoGet.CombinedOutput())(func(data map[string]interface{}) (interface{}, error) {
		if len(data) > 0 {
			switch strings.ToUpper(data["level"].(string)) {
			case "INFO", "DEBUG":
				return data["file"], nil
			case "ERROR":
				return nil, errors.New(data["message"].(string))
			}
		}
		return nil, unknownReturnStatus
	})

	if result != nil {
		file = result.(string)
	}
	return
}

func (p *proxmoxProvider) getImage(name string, images []*Image) (*Image, error) {
	for _, image := range images {
		if strings.ToLower(name) == strings.ToLower(image.Name) {
			return image, nil
		}
	}
	return nil, fmt.Errorf("no image by name %s", name)
}

// ---------------------------------------------------------------------------------------------------------------------

const (
	keyInfra = "infra"
	keyName  = "name"
	proxmox  = "proxmox"
	tempDir  = "/tmp"
)

var (
	unknownReturnStatus = errors.New("unknown_return_status")
)
