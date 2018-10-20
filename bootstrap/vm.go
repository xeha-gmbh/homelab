package bootstrap

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// main entry point to create VM structures from YAML file
func ParseVMs(data map[string]interface{}) ([]*VM, error) {
	rawVMs, isList := data[keyVMs].([]interface{})
	if !isList {
		output.Fatal(1,
			"Malformed config: {{index .error}}",
			map[string]interface{}{
				"event":    "parse_error",
				"exitCode": 1,
				"error":    fmt.Sprintf("expect key '%s' to be a list", keyVMs),
			})
		return nil, errors.New("parse_error")
	}

	vms := make([]*VM, 0, len(rawVMs))
	for _, oneRawVM := range rawVMs {
		rawData, isMap := oneRawVM.(map[interface{}]interface{})
		if !isMap {
			output.Fatal(1,
				"Malformed config: {{index .error}}",
				map[string]interface{}{
					"event":    "parse_error",
					"exitCode": 1,
					"error": fmt.Sprintf("expect each '%s' to be a map, but got %s",
						keyVMs, reflect.TypeOf(oneRawVM).String()),
				})
			return nil, errors.New("parse_error")
		}

		vm := &VM{}
		if err := mapstructure.Decode(rawData, vm); err != nil {
			output.Fatal(1,
				"Malformed config: unable to decode vm. Cause: {{index .cause}}",
				map[string]interface{}{
					"event":    "parse_error",
					"exitCode": 1,
					"cause":    err.Error(),
				})
			return nil, errors.New("parse_error")
		}

		switch vm.Provider.Name {
		case proxmox:
			if vm.Archetype == basicArchetype {
				params, err := ParseProxmoxBasicArchetypeParams(rawData["params"])
				if err != nil {
					output.Fatal(1,
						"Malformed config: unable to parse proxmox basic params. Cause: {{index .cause}}",
						map[string]interface{}{
							"event":    "parse_error",
							"exitCode": 1,
							"cause":    err.Error(),
						})
					return nil, errors.New("parse_error")
				}
				vm.Params = params
			} else {
				output.Fatal(1,
					"Unsupported proxmox archetype {{index .archetype}}.",
					map[string]interface{}{
						"event":     "api_error",
						"exitCode":  1,
						"archetype": vm.Archetype,
					})
				return nil, errors.New("api_error")
			}
		default:
			output.Fatal(1,
				"Unsupported provider {{index .provider}}.",
				map[string]interface{}{
					"event":    "api_error",
					"exitCode": 1,
					"provider": vm.Provider.Name,
				})
			return nil, errors.New("api_error")
		}

		vms = append(vms, vm)
	}

	return vms, nil
}

type VM struct {
	Id       string `yaml:"id"`
	Name     string `yaml:"name"`
	Provider struct {
		Name string                 `yaml:"name"`
		Args map[string]interface{} `yaml:"args"`
	} `yaml:"provider"`
	Image struct {
		Name  string `yaml:"name"`
		Store string `yaml:"store"`
	} `yaml:"image"`
	Archetype string      `yaml:"archetype"`
	Params    interface{} `yaml:"-"`
	Start     bool        `yaml:"start"`
}

// ---------------------------------------------------------------------------------------------------------------------

func ParseProxmoxBasicArchetypeParams(data interface{}) (*proxmoxBasicArchetypeParams, error) {
	p := new(proxmoxBasicArchetypeParams)
	if err := mapstructure.Decode(data, p); err != nil {
		return nil, fmt.Errorf("failed to parse proxmox basic params: %s", err.Error())
	}

	if ok, err := regexp.MatchString("^\\d+[MmGg]$", p.Memory); err != nil || !ok {
		return nil, fmt.Errorf("malformed memory size %s", p.Memory)
	}

	if ok, err := regexp.MatchString("^\\d+[MmGg]$", p.Drive.Size); err != nil || !ok {
		return nil, fmt.Errorf("malformed drive size %s", p.Drive.Size)
	}

	ips := []string{p.Network.Ip, p.Network.Mask, p.Network.Gateway}
	ips = append(ips, p.Network.Dns...)
	for _, ip := range ips {
		if ok, err := regexp.MatchString("^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}$", ip); err != nil || !ok {
			return nil, fmt.Errorf("malformed ip address %s", ip)
		}
	}

	return p, nil
}

type proxmoxBasicArchetypeParams struct {
	Cpu    int    `yaml:"cpu"`
	Memory string `yaml:"memory"`
	Drive  struct {
		Store string `yaml:"store"`
		Size  string `yaml:"size"`
	} `yaml:"drive"`
	Network struct {
		Interface string   `yaml:"interface"`
		Ip        string   `yaml:"ip"`
		Mask      string   `yaml:"mask"`
		Gateway   string   `yaml:"gateway"`
		Dns       []string `yaml:"dns"`
	} `yaml:"network"`
	System struct {
		Timezone string `yaml:"timezone"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Hostname string `yaml:"hostname"`
		Domain   string `yaml:"domain"`
	} `yaml:"system"`
}

func (p *proxmoxBasicArchetypeParams) MemoryMB() int {
	amount, unit, err := p.amountAndUnit(p.Memory)
	if err != nil {
		panic("invalid state: memory size not a number")
	}

	switch strings.ToUpper(unit) {
	case "M":
		return amount
	case "G":
		return amount * 1024
	default:
		panic("invalid state: unsupported memory unit")
	}
}

func (p *proxmoxBasicArchetypeParams) DriveGB() int {
	amount, unit, err := p.amountAndUnit(p.Drive.Size)
	if err != nil {
		panic("invalid state: drive size not a number")
	}

	switch strings.ToUpper(unit) {
	case "M":
		return amount / 1024
	case "G":
		return amount
	default:
		panic("invalid state: unsupported drive size unit")
	}
}

func (p *proxmoxBasicArchetypeParams) amountAndUnit(value string) (int, string, error) {
	amount, unit := value[:len(value)-1], value[len(value)-1:]
	i, err := strconv.Atoi(amount)
	if err != nil {
		return 0, "", err
	}
	return i, unit, nil
}

// ---------------------------------------------------------------------------------------------------------------------

const (
	keyVMs         = "vms"
	basicArchetype = "basic"
)
