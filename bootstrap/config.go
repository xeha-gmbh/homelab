package bootstrap

import (
	"github.com/imulab/homelab/shared"
	"gopkg.in/yaml.v2"
	"os"
)

func ParseConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		output.Error(shared.ErrParse.ExitCode,
			"Unable to open file {{index .file}}. Cause: {{index .cause}}",
			map[string]interface{}{
				"event": "parse_error",
				"file": path,
				"cause": err.Error(),
			})
		return nil, shared.ErrParse
	}

	raw := make(map[string]interface{})
	err = yaml.NewDecoder(f).Decode(&raw)
	if err != nil {
		output.Error(shared.ErrParse.ExitCode,
			"Unable to parse file {{index .file}}. Cause: {{index .cause}}",
			map[string]interface{}{
				"event": "parse_error",
				"file": path,
				"cause": err.Error(),
			})
		return nil, shared.ErrParse
	}

	switch raw["version"].(string) {
	case "1":
		return parseV1Config(raw)
	default:
		output.Error(shared.ErrApi.ExitCode,
			"Unsupported API version {{index .version}}",
			map[string]interface{}{
				"event": "api_error",
				"version": raw["version"].(string),
			})
		return nil, shared.ErrApi
	}
}

type Config interface {
	Bootstrap() error
}
