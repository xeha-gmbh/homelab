package login

import (
	"github.com/imulab/homelab/shared"
)

var (
	ErrAuth = shared.ErrorFactory(10)("authentication_error")
)
