package bootstrap

import (
	"errors"
	"fmt"
	"github.com/imulab/homelab/shared"
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

	if dlImagePath, err = p.ensureImage(vm, image); err != nil {
		return err
	}

	if autoImagePath, err = p.createAutoInstallImage(vm, image, dlImagePath); err != nil {
		return err
	}

	fmt.Println(autoImagePath)

	return nil
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
			"--name-servers", strings.Join(params.Network.Dns, ","),
		}...)
	default:
		return "", fmt.Errorf("unknown archetype %s", vm.Archetype)
	}
	isoAuto := exec.Command("homelab", isoAutoArgs...)

	result, err := shared.HandledJson(isoAuto.CombinedOutput())(func(data map[string]interface{}) (interface{}, error) {
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

	result, err := shared.HandledJson(isoGet.CombinedOutput())(func(data map[string]interface{}) (interface{}, error) {
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
