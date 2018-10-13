package login

import (
	"fmt"
	"os"
)

const (
	GenericErrorExitCode        = 1
	ProxmoxErrorExitCode        = 2
	ValidationErrorExitCode     = 3
	AuthenticationErrorExitCode = 4
)

type exitCodeAwareError interface {
	error
	ExitCode() int
}

func handleError(err error) error {
	switch err.(type) {
	case exitCodeAwareError:
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(err.(exitCodeAwareError).ExitCode())
	default:
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(GenericErrorExitCode)
	}
	return err
}

type commandError struct {
	Err  error
	Code int
}

func (e *commandError) Error() string {
	return e.Err.Error()
}

func (e *commandError) ExitCode() int {
	return e.Code
}

func genericError(err error) error {
	return &commandError{
		Err:  err,
		Code: GenericErrorExitCode,
	}
}

func proxmoxError(err error) error {
	return &commandError{
		Err:  err,
		Code: ProxmoxErrorExitCode,
	}
}

func authenticationError(err error) error {
	return &commandError{
		Err:  err,
		Code: AuthenticationErrorExitCode,
	}
}
