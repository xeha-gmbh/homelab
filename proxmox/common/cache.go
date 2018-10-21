package common

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

const (
	TicketCache = ".proxmox"
)

// Session information representing an authenticated Proxmox user
type ProxmoxSubject struct {
	Username  string `json:"username"`
	Ticket    string `json:"ticket"`
	CSRFToken string `json:"csrf_token"`
	ApiServer string `json:"api_server"`
}

// Read session subject from ticket cache
func ReadSubjectFromCache() (*ProxmoxSubject, error) {
	var (
		err     error
		file    *os.File
		b       []byte
		subject ProxmoxSubject
	)

	file, err = os.Open(proxmoxTicketCache())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	b, err = ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &subject)
	if err != nil {
		return nil, err
	}

	return &subject, nil
}

// Write session information to ticket cache in (pretty) JSON format.
func WriteSubjectToCache(subject *ProxmoxSubject) error {
	ticketCachePath := proxmoxTicketCache()

	if b, err := json.MarshalIndent(subject, "", "    "); err != nil {
		return err
	} else if err := ioutil.WriteFile(ticketCachePath, b, 0600); err != nil {
		return err
	}

	return nil
}

// Returns the expected Proxmox ticket cache location for the current user.
// If current user home directory cannot be acquired, it defaults to current directory.
func proxmoxTicketCache() string {
	if u, err := user.Current(); err != nil {
		return TicketCache
	} else {
		return filepath.Join(u.HomeDir, TicketCache)
	}
}
