package common

import (
	"fmt"
	"os"
)

type ExitAwareError interface {
	error
	ExitStatus() int
}

func HandleError(err error) error {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		switch err.(type) {
		case ExitAwareError:
			os.Exit(err.(ExitAwareError).ExitStatus())
		}
	}
	return err
}
