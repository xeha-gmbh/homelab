package bootstrap

import (
	"fmt"
	. "github.com/imulab/homelab/shared"
	"strings"
)

func parseV1Config(data map[string]interface{}) (Config, error) {
	providers, err := ParseProviders(data)
	if err != nil {
		return nil, err
	}

	images, err := ParseImages(data)
	if err != nil {
		return nil, err
	}

	vms, err := ParseVMs(data)
	if err != nil {
		return nil, err
	}

	return &v1Config{Providers: providers, Images: images, VMs: vms}, nil
}

type v1Config struct {
	Providers 	[]Provider		`yaml:"providers"`
	Images 		[]*Image 		`yaml:"images"`
	VMs 		[]*VM			`yaml:"vms"`
	out 		MessagePrinter	`yaml:"-"`
}

func (c *v1Config) Bootstrap() error {
	//yaml.NewEncoder(os.Stdout).Encode(c)
	//yaml.NewEncoder(os.Stdout).Encode(c.VMs[0].Params)

	for _, vm := range c.VMs {
		provider, err := c.GetProvider(vm.Provider.Name)
		if err != nil {
			output.Fatal(ErrOp.ExitCode,
				"Failed to creating vm [name={{index .name}}]. Cause: {{index .cause}}.",
				map[string]interface{}{
					"event": "",
					"name": vm.Name,
					"cause": err.Error(),
				})
			return ErrOp
		}

		err = provider.CreateVM(vm, c.Images)
		if err != nil {
			output.Fatal(ErrOp.ExitCode,
				"Failed to creating vm [name={{index .name}}]. Cause: {{index .cause}}.",
				map[string]interface{}{
					"event": "",
					"name": vm.Name,
					"cause": err.Error(),
				})
			return ErrOp
		}
	}

	return nil
}

func (c *v1Config) GetProvider(name string) (Provider, error) {
	for _, provider := range c.Providers {
		if strings.ToLower(name) == strings.ToLower(provider.Name()) {
			return provider, nil
		}
	}
	return nil, fmt.Errorf("no provider by name %s", name)
}