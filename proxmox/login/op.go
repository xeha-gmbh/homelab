package login

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

// Arguments for the 'proxmox login' command
type ProxmoxLoginRequest struct {
	Username  string
	Password  string
	Realm     string
	ApiServer string
	Force     bool
}

// Performs a login using the parameters supplied. This method only performs a new login attempt
// when ProxmoxLoginRequest#Force is set to true, or a ticket cache cannot be found or used.
func (pl *ProxmoxLoginRequest) Login() (*ProxmoxSubject, bool, error) {
	if cachedSubject, err := NewProxmoxSubjectFromFile(proxmoxTicketCache()); err != nil || pl.Force {
		s, e := pl.doLogin()
		return s, true, e
	} else {
		fmt.Fprintf(os.Stdout, "Ticket exists in cache.\n")
		return cachedSubject, false, nil
	}
}

// Performs a real login attempt. This method returns error when
// 1) HTTP request returns error
// 2) Proxmox returns status 401 (special case)
// 3) Proxmox returns other non-200 status
// 4) Response body cannot be decoded properly
// Otherwise, it returns a nil error and a ProxmoxSubject
func (pl *ProxmoxLoginRequest) doLogin() (*ProxmoxSubject, error) {
	var (
		err    error
		resp   *http.Response
		client = httpClient()
		form   = url.Values{}
	)

	form.Add("username", pl.Username)
	form.Add("password", pl.Password)
	form.Add("realm", pl.Realm)

	resp, err = client.PostForm(loginUrl(pl.ApiServer), form)
	if err != nil {
		return nil, proxmoxError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, authenticationError(errors.New("authentication failure"))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, proxmoxError(errors.New("request failure"))
	}

	respData := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return nil, genericError(err)
	}

	subject := &ProxmoxSubject{
		Username:  respData["data"].(map[string]interface{})["username"].(string),
		CSRFToken: respData["data"].(map[string]interface{})["CSRFPreventionToken"].(string),
		Ticket:    respData["data"].(map[string]interface{})["ticket"].(string),
	}

	fmt.Fprintf(os.Stdout, "User '%s' successfully logged in.\n", subject.Username)
	return subject, nil
}

// Session information representing an authenticated Proxmox user
type ProxmoxSubject struct {
	Username  string `json:"username"`
	Ticket    string `json:"ticket"`
	CSRFToken string `json:"csrf_token"`
}

// Write session information to file at given path in (pretty) JSON format.
func (ps *ProxmoxSubject) WriteToFile(filePath string) error {
	if b, err := json.MarshalIndent(ps, "", "    "); err != nil {
		return genericError(err)
	} else if err := ioutil.WriteFile(filePath, b, 0600); err != nil {
		return genericError(err)
	}

	fmt.Fprintln(os.Stdout, "Ticket cache written to", filePath)
	return nil
}

// Read session information from JSON file at given path. This method returns error when
// 1) File cannot be opened
// 2) Read file returns error
// 3) File cannot be parsed as JSON
// Otherwise, it returns nil error and a ProxmoxSubject.
func NewProxmoxSubjectFromFile(filePath string) (*ProxmoxSubject, error) {
	var (
		err     error
		file    *os.File
		b       []byte
		subject ProxmoxSubject
	)

	file, err = os.Open(filePath)
	if err != nil {
		return nil, genericError(err)
	}
	defer file.Close()

	b, err = ioutil.ReadAll(file)
	if err != nil {
		return nil, genericError(err)
	}

	err = json.Unmarshal(b, &subject)
	if err != nil {
		return nil, genericError(err)
	}

	return &subject, nil
}
