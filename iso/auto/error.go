package auto

import (
	"fmt"
	"os"
	"strings"
)

const (
	errorGeneric = "generic_error"
	errorNoProvider = "no_provider"
)

func handleError(err error) error {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())

		switch {
		case strings.HasPrefix(err.Error(), errorGeneric):
			os.Exit(1)
		default:
			os.Exit(255)
		}
	}

	return err
}

func genericError(reason string) error {
	return fmt.Errorf("%s:%s", errorGeneric, reason)
}
