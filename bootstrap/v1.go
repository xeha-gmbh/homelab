package bootstrap

import (
	"gopkg.in/yaml.v2"
	"os"
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
	Providers 	[]Provider	`yaml:"providers"`
	Images 		[]*Image 	`yaml:"images"`
	VMs 		[]*VM		`yaml:"vms"`
}

func (c *v1Config) Bootstrap() error {
	yaml.NewEncoder(os.Stdout).Encode(c)
	//yaml.NewEncoder(os.Stdout).Encode(c.VMs[0].Params)
	return nil
}
