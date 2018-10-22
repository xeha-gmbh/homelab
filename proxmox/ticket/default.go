package ticket

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"os"
	"os/user"
	"path/filepath"
	"sync"
)

var (
	oneDefaultService      sync.Once
	defaultServiceInstance Service
)

// Main entry point for default service
func DefaultService() Service {
	oneDefaultService.Do(func() {
		defaultServiceInstance = newDefaultService()
	})
	return defaultServiceInstance
}

// Main constructor for defaultService
func newDefaultService() Service {
	return &defaultService{
		dotProxmox: ".proxmox",
	}
}

// defaultService implements Service interface
type defaultService struct {
	dotProxmox string
}

func (s *defaultService) Get() (*Session, error) {
	return s.GetFrom(s.DefaultStorage())
}

func (s *defaultService) GetFrom(path string) (*Session, error) {
	var err error

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	session := &Session{}
	if err = json.NewDecoder(f).Decode(session); err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"storage": path,
	}).Debug("read session from storage.")
	return session, nil
}

func (s *defaultService) Save(session *Session) error {
	return s.SaveTo(session, s.DefaultStorage())
}

func (s *defaultService) SaveTo(session *Session, path string) error {
	var err error

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	if err = json.NewEncoder(f).Encode(session); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"storage": path,
	}).Debug("saved session to storage.")
	return nil
}

// Storage is '.proxmox' file in the base directory which defaults user's home directory. If current user cannot be
// identified, base directory falls back to current directory.
func (s *defaultService) DefaultStorage() string {
	u, err := user.Current()
	if err != nil {
		return s.dotProxmox
	}
	return filepath.Join(u.HomeDir, s.dotProxmox)
}
