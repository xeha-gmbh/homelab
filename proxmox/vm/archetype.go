package vm

import (
	flag "github.com/spf13/pflag"
	"sync"
)

// list archetype command
// each payload is specific to an archetype, but shares shared logic -> interface

// Description of a collection of VM configuration options.
// It is designed to ease the VM creation process by only exposing a small amount of configurable
// options while fixing most options. Each archetype will be different, so this interface is also
// expected to be an adapter to cobra.Command in order to supply command creation information.
type Archetype interface {
	// Returns the name of the archetype
	Name() string
	// Returns the name of the sub-command under 'lab proxmox vm create', must be unique.
	Use() string
	// Returns the basic description of this archetype
	Short() string
	// Returns the long description of this archetype
	Long() string
	// Returns the list of all available configuration flags
	AllFlags() []string
	// Returns the list of required flags for the 'lab proxmox vm create ${archetype.name}' command
	RequiredFlags() []string
	// Bind flags to its internal payload
	BindFlags(flagSet *flag.FlagSet)
	// Create VM
	CreateVM() error
}

var (
	oneArchetypeRepo 	sync.Once
	archetypeRepo		*archetypeRepository
)

// Public entry point of accessing the archetype repository
func ArchetypeRepository() *archetypeRepository {
	oneArchetypeRepo.Do(func() {
		archetypeRepo = &archetypeRepository{
			at: make(map[string]Archetype),
		}
	})
	return archetypeRepo
}

// Repository for storing all available archetypes
type archetypeRepository struct {
	at 	map[string]Archetype
}

// Submits an archetype to the repository. Returns true if the the archetype
// is accepted and will be available as a command. Return false if an archetype
// with the same Archetype#Use already exists.
func (r *archetypeRepository) SubmitArchetype(t Archetype) bool {
	if _, ok := r.at[t.Use()]; !ok {
		r.at[t.Use()] = t
		return true
	}
	return false
}

// Returns a copy of all archetypes.
func (r *archetypeRepository) AllArchetypes() []Archetype {
	all := make([]Archetype, 0, len(r.at))
	for _, v := range r.at {
		all = append(all, v)
	}
	return all
}

// configuration data for a command line argument
type Arg struct {
	Name 		string
	Default 	interface{}
	Description	string
}