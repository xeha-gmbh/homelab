package login

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/imulab/homelab/proxmox/common"
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
func (pl *ProxmoxLoginRequest) Login() (*common.ProxmoxSubject, bool, error) {
	if cachedSubject, err := common.ReadSubjectFromCache(); err != nil || pl.Force {
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
func (pl *ProxmoxLoginRequest) doLogin() (*common.ProxmoxSubject, error) {
	var (
		err    error
		resp   *http.Response
		client = common.HttpClient()
		form   = url.Values{}
	)

	form.Add("username", pl.Username)
	form.Add("password", pl.Password)
	form.Add("realm", pl.Realm)

	resp, err = client.PostForm(loginUrl(pl.ApiServer), form)
	if err != nil {
		return nil, common.ProxmoxError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, authenticationError(errors.New("authentication failure"))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, common.ProxmoxError(errors.New("request failure"))
	}

	respData := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return nil, common.GenericError(err)
	}

	subject := &common.ProxmoxSubject{
		Username:  respData["data"].(map[string]interface{})["username"].(string),
		CSRFToken: respData["data"].(map[string]interface{})["CSRFPreventionToken"].(string),
		Ticket:    respData["data"].(map[string]interface{})["ticket"].(string),
		ApiServer: pl.ApiServer,
	}

	fmt.Fprintf(os.Stdout, "User '%s' successfully logged in.\n", subject.Username)
	return subject, nil
}

// Returns the ticket API url for the given Proxmox host.
func loginUrl(base string) string {
	return fmt.Sprintf("%s/api2/json/access/ticket", base)
}
