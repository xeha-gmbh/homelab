package login

import (
	"github.com/xeha-gmbh/homelab/shared"
)

var (
	ErrAuth = shared.ErrorFactory(10)("authentication_error")
)
