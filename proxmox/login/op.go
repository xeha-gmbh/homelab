package login

import (
	"encoding/json"
	"fmt"
	"github.com/xeha-gmbh/homelab/proxmox/common"
	"github.com/xeha-gmbh/homelab/shared"
	"net/http"
	"net/url"
)

// Arguments for the 'proxmox login' command
type ProxmoxLoginRequest struct {
	shared.ExtraArgs
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
		output.Info("Ticket exists in cache.", map[string]interface{}{})
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

	output.Debug("Login request return status code {{index .code}}",
		map[string]interface{}{
			"event":  "login_response",
			"code":   resp.StatusCode,
			"status": resp.Status,
		})

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrAuth
	}

	if resp.StatusCode != http.StatusOK {
		return nil, ErrAuth
	}

	respData := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return nil, shared.ErrParse
	}

	subject := &common.ProxmoxSubject{
		Username:  respData["data"].(map[string]interface{})["username"].(string),
		CSRFToken: respData["data"].(map[string]interface{})["CSRFPreventionToken"].(string),
		Ticket:    respData["data"].(map[string]interface{})["ticket"].(string),
		ApiServer: pl.ApiServer,
	}

	output.Info("User {{index .user}} is now logged in.",
		map[string]interface{}{
			"event": "login_success",
			"user":  subject.Username,
		})
	return subject, nil
}

// Returns the ticket API url for the given Proxmox host.
func loginUrl(base string) string {
	return fmt.Sprintf("%s/api2/json/access/ticket", base)
}
