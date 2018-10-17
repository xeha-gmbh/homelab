package shared

// Provider for creating unattended iso.
type Provider interface {
	// Returns true if the supplied flavor is supported by this provider for processing.
	SupportsFlavor(flavor string) bool
	// Returns true if all the underlying package dependencies have been met.
	CheckDependencies(payload *Payload) (bool, error)
	// Process the ISO to create an unattended installation media
	RemasterISO(payload *Payload) error
}

type NoProviderError struct{}

func (p NoProviderError) Error() string {
	return "no provider to handle flavor"
}

func (p NoProviderError) ExitStatus() int {
	return 2
}
