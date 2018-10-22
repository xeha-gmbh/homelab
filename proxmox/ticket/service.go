package ticket

// Service interface for Proxmox session ticket I/O.
type Service interface {

	// Retrieve the stored session from default location.
	Get() (*Session, error)

	// Retrieve the stored session from user specified location
	GetFrom(path string) (*Session, error)

	// Save the session to default location, overrides any existing ones.
	Save(*Session) error

	// Save the session to user specified location, overrides any existing ones.
	SaveTo(*Session, string) error

	// Default session storage location
	DefaultStorage() string
}
