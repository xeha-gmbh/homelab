package bootstrap

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

func ParseConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %s", path, err.Error())
	}

	raw := make(map[string]interface{})
	err = yaml.NewDecoder(f).Decode(&raw)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %s: %s", path, err.Error())
	}

	switch raw["version"].(string) {
	case "1":
		return parseV1Config(raw)
	default:
		return nil, fmt.Errorf("api version %s not supported", raw["version"].(string))
	}
}

type Config interface {
	Bootstrap() error
}
