package login

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os/user"
	"path/filepath"
)

// Returns an http client.
// TODO: support secure transport
func httpClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return &http.Client{
		Transport: tr,
	}
}

// Returns the ticket API url for the given Proxmox host.
func loginUrl(base string) string {
	return fmt.Sprintf("%s/api2/json/access/ticket", base)
}

// Returns the expected Proxmox ticket cache location for the current user.
// If current user home directory cannot be acquired, it defaults to current directory.
func proxmoxTicketCache() string {
	if u, err := user.Current(); err != nil {
		handleError(&commandError{
			Err:  err,
			Code: GenericErrorExitCode,
		})
		return TicketCache
	} else {
		return filepath.Join(u.HomeDir, TicketCache)
	}
}
