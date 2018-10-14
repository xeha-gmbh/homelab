package common

import (
	"fmt"
	"os"
)

const (
	GenericErrorExitCode        = 1
	ProxmoxErrorExitCode        = 2
	ValidationErrorExitCode     = 3
)

type ExitCodeAwareError interface {
	error
	ExitCode() int
}

func HandleError(err error) error {
	switch err.(type) {
	case ExitCodeAwareError:
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(err.(ExitCodeAwareError).ExitCode())
	default:
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(GenericErrorExitCode)
	}
	return err
}

type CommandError struct {
	Err  error
	Code int
}

func (e *CommandError) Error() string {
	return e.Err.Error()
}

func (e *CommandError) ExitCode() int {
	return e.Code
}

func GenericError(err error) error {
	return &CommandError{
		Err:  err,
		Code: GenericErrorExitCode,
	}
}

func ProxmoxError(err error) error {
	return &CommandError{
		Err:  err,
		Code: ProxmoxErrorExitCode,
	}
}


