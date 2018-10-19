package bootstrap

import (
	"errors"
	"fmt"
	"github.com/imulab/homelab/iso/get"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"strings"
)

// Entry point to parse a list of providers.
// Input data expects a top level key whose name is the value of 'keyInfra'
func ParseProviders(data map[string]interface{}) ([]Provider, error) {
	rawProviders, isList := data[keyInfra].([]interface{})
	if !isList {
		return nil, fmt.Errorf("expect key '%s' to be a list", keyInfra)
	}

	providers := make([]Provider, 0, len(rawProviders))
	for _, oneRawProvider := range rawProviders {
		rawData, isMap := oneRawProvider.(map[interface{}]interface{})
		if !isMap {
			return nil, fmt.Errorf("expect each '%s' to be a map, but got %s",
				keyInfra, reflect.TypeOf(oneRawProvider).String())
		}
		providerName, hasName := rawData[keyName].(string)
		if !hasName {
			return nil, fmt.Errorf("expect each '%s' to have a key '%s'", keyInfra, keyName)
		}

		var oneProvider Provider
		switch providerName {
		case proxmox:
			oneProvider = &proxmoxProvider{}
			if err := mapstructure.Decode(rawData, oneProvider); err != nil {
				return nil, fmt.Errorf("decode proxmox provider failed: %s", err.Error())
			}
		default:
			return nil, fmt.Errorf("unsupported provider %s", providerName)
		}
		providers = append(providers, oneProvider)
	}

	if len(providers) == 0 {
		return nil, errors.New("no providers")
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
	if err := p.ensureImage(vm, images); err != nil {
		return err
	}

	return nil
}

func (p *proxmoxProvider) ensureImage(vm *VM, images []*Image) error {
	image, err := p.getImage(vm.Image.Name, images)
	if err != nil {
		return err
	}

	isoGet := get.NewIsoGetCommand()
	isoGet.SetArgs([]string{
		"--flavor",	image.Flavor,
		"--target-dir", tempDir,
		"--reuse",
	})
	return isoGet.Execute()
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
