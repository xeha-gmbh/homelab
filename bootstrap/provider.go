package bootstrap

import (
	"errors"
	"fmt"
	"github.com/imulab/homelab/shared"
	"github.com/mitchellh/mapstructure"
	"os/exec"
	"reflect"
	"strings"
)

// Entry point to parse a list of providers.
// Input data expects a top level key whose name is the value of 'keyInfra'
func ParseProviders(data map[string]interface{}) ([]Provider, error) {
	rawProviders, isList := data[keyInfra].([]interface{})
	if !isList {
		output.Error(shared.ErrParse.ExitCode,
			"Malformed config: {{index .error}}",
			map[string]interface{}{
				"event": "parse_error",
				"error": fmt.Sprintf("expect key '%s' to be a list.", keyInfra),
			})
		return nil, shared.ErrParse
	}

	providers := make([]Provider, 0, len(rawProviders))
	for _, oneRawProvider := range rawProviders {
		rawData, isMap := oneRawProvider.(map[interface{}]interface{})
		if !isMap {
			output.Error(shared.ErrParse.ExitCode,
				"Malformed config: {{index .error}}",
				map[string]interface{}{
					"event": "parse_error",
					"error": fmt.Sprintf("expect each '%s' to be a map, but got %s",
						keyInfra, reflect.TypeOf(oneRawProvider).String()),
				})
			return nil, shared.ErrParse
		}
		providerName, hasName := rawData[keyName].(string)
		if !hasName {
			output.Error(shared.ErrParse.ExitCode,
				"Malformed config: {{index .error}}",
				map[string]interface{}{
					"event": "parse_error",
					"error": fmt.Sprintf("expect each '%s' to have a key '%s'", keyInfra, keyName),
				})
			return nil, shared.ErrParse
		}

		var oneProvider Provider
		switch providerName {
		case proxmox:
			oneProvider = &proxmoxProvider{}
			if err := mapstructure.Decode(rawData, oneProvider); err != nil {
				output.Error(shared.ErrParse.ExitCode,
					"Malformed config, unable to decode provider. Cause: {{index .cause}}",
					map[string]interface{}{
						"event": "parse_error",
						"cause": err.Error(),
					})
				return nil, shared.ErrParse
			}
		default:
			output.Error(shared.ErrApi.ExitCode,
				"Unsupported provider {{index .provider}}.",
				map[string]interface{}{
					"event": "api_error",
					"provider": providerName,
				})
			return nil, shared.ErrApi
		}
		providers = append(providers, oneProvider)
	}

	if len(providers) == 0 {
		output.Error(shared.ErrApi.ExitCode,
			"No provider.",
			map[string]interface{}{
				"event": "api_error",
			})
		return nil, shared.ErrApi
	}

	return providers, nil
}
// ---------------------------------------------------------------------------------------------------------------------

// Interface for all providers
type Provider interface {
	Name() string
	CreateVM(vm *VM, images []*Image) error
}
// ---------------------------------------------------------------------------------------------------------------------

// The proxmox provider
type proxmoxProvider struct {
	Api 		string 		`yaml:"api"`
	Identity	struct{
		Realm		string		`yaml:"realm"`
		Username	string		`yaml:"username"`
		Password 	string 		`yaml:"password"`
	}						`yaml:"identity"`
	DataStores	[]struct{
		Name 		string		`yaml:"name"`
		Tags		[]string 	`yaml:"tags"`
	}						`yaml:"datastores"`
}

func (p *proxmoxProvider) Name() string {
	return proxmox
}

func (p *proxmoxProvider) CreateVM(vm *VM, images []*Image) error {
	if file, err := p.ensureImage(vm, images); err != nil {
		return err
	} else {
		fmt.Println(file)
	}

	return nil
}

func (p *proxmoxProvider) ensureImage(vm *VM, images []*Image) (file string, err error) {
	image, err := p.getImage(vm.Image.Name, images)
	if err != nil {
		return
	}

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
		return nil, errors.New("unknown_return_status")
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
	keyName = "name"
	proxmox  = "proxmox"

	tempDir	= "/tmp"
)
