package auto

// Provider for creating unattended iso.
type Provider interface {
	// Returns true if the supplied flavor is supported by this provider for processing.
	SupportsFlavor(flavor string) bool
	// Returns true if all the underlying package dependencies have been met.
	CheckDependencies() (bool, error)
	// Process the ISO to create an unattended installation media
	RemasterISO(payload *Payload) error
}