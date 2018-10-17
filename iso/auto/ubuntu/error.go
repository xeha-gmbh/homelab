package ubuntu

import (
	"fmt"
)

// Generic
type GenericError struct {
	Reason	string
}

func (e GenericError) Error() string {
	return e.Reason
}

func (e GenericError) ExitStatus() int {
	return 1
}

func NewGenericError(reason string) error {
	return &GenericError{Reason: reason}
}

// Dependency
type DependencyError struct {
	MissingPkg		string
	SuggestedPkg	string
}

func (e DependencyError) Error() string {
	return fmt.Sprintf("Package %s not found. Try install %s", e.MissingPkg, e.SuggestedPkg)
}

func (e DependencyError) ExitStatus() int {
	return 4
}

func NewDependencyError(missing, suggest string) error {
	return &DependencyError{MissingPkg:missing, SuggestedPkg:suggest}
}