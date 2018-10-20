package bootstrap

import (
	"fmt"
	"github.com/imulab/homelab/shared"
	"github.com/mitchellh/mapstructure"
	"reflect"
)

// Entry point to parse a list of providers.
// Input data expects a top level key whose name is the value of 'keyInfra'
func ParseProviders(data map[string]interface{}) ([]Provider, error) {
	rawProviders, isList := data[keyInfra].([]interface{})
	if !isList {
		output.Fatal(shared.ErrParse.ExitCode,
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
			output.Fatal(shared.ErrParse.ExitCode,
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
			output.Fatal(shared.ErrParse.ExitCode,
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
				output.Fatal(shared.ErrParse.ExitCode,
					"Malformed config, unable to decode provider. Cause: {{index .cause}}",
					map[string]interface{}{
						"event": "parse_error",
						"cause": err.Error(),
					})
				return nil, shared.ErrParse
			}
		default:
			output.Fatal(shared.ErrApi.ExitCode,
				"Unsupported provider {{index .provider}}.",
				map[string]interface{}{
					"event":    "api_error",
					"provider": providerName,
				})
			return nil, shared.ErrApi
		}
		providers = append(providers, oneProvider)
	}

	if len(providers) == 0 {
		output.Fatal(shared.ErrApi.ExitCode,
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
