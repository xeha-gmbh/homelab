package auto

import (
	"fmt"
	"strings"
)

const (
	flavorUbuntuBionic64 = "ubuntu/bionic64"
	flavorUbuntuXenial64 = "ubuntu/xenial64"
)

func init() {
	allProviders = append(allProviders, &UbuntuProvider{})
}

type UbuntuProvider struct {}

func (p *UbuntuProvider) SupportsFlavor(flavor string) bool {
	switch strings.ToLower(flavor) {
	case flavorUbuntuBionic64, flavorUbuntuXenial64:
		return true
	default:
		return false
	}
}

func (p *UbuntuProvider) CheckDependencies() (bool, error) {
	return true, nil
}

func (p *UbuntuProvider) RemasterISO(payload *Payload) error {
	fmt.Println(payload.Flavor, payload.Username, payload.Password)
	return nil
}

