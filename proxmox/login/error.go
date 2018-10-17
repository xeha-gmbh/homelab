package login

import "github.com/imulab/homelab/proxmox/common"

const (
	AuthenticationErrorExitCode = 10
)

func authenticationError(err error) error {
	return &common.CommandError{
		Err:  err,
		Code: AuthenticationErrorExitCode,
	}
}
