package common

import (
	"crypto/tls"
	"net/http"
)

const (
	ProxmoxCSRFTokenHeader  = "CSRFPreventionToken"
	ProxmoxTicketCookieName = "PVEAuthCookie"
)

func WithHttpCredentials(r *http.Request) (*http.Request, error) {
	if subject, err := ReadSubjectFromCache(); err != nil {
		return r, err
	} else {
		r.Header.Set(ProxmoxCSRFTokenHeader, subject.CSRFToken)

		cookie := &http.Cookie{
			Name:  ProxmoxTicketCookieName,
			Value: subject.Ticket,
		}
		r.AddCookie(cookie)

		return r, nil
	}
}

// Returns an http client.
// TODO: support secure transport
func HttpClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return &http.Client{
		Transport: tr,
	}
}
