package upload

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/imulab/homelab/proxmox/common"
	"github.com/imulab/homelab/shared"
	"net/http"
	"os/exec"
	"strings"
)

// Arguments for 'proxmox upload' command.
type ProxmoxUploadRequest struct {
	shared.ExtraArgs
	Node    string
	Storage string
	File    string
	Format  string
}

// Perform upload. If ProxmoxUploadRequest#Storage is not set, this method will try to
// query the Proxmox API for the first storage device that accepts ProxmoxUploadRequest#Format
// and use that device as the storage option.
func (ur *ProxmoxUploadRequest) Upload() error {
	var err error

	if len(strings.TrimSpace(ur.Storage)) == 0 {
		if ur.Storage, err = ur.matchFirstStorageDevice(); err != nil {
			return err
		}
	}

	return ur.doUpload()
}

// Actually perform the upload operation
// For unknown reason, HTTP multipart support in Golang does not play well with Proxmox API.
// Hence, we defer to using curl to perform the web request here.
func (ur *ProxmoxUploadRequest) doUpload() error {
	if subject, err := common.ReadSubjectFromCache(); err != nil {
		return common.GenericError(fmt.Errorf("failed to read ticket cache: %s", err.Error()))
	} else {
		curl := exec.Command("curl", "-k",
			"-H", fmt.Sprintf("CSRFPreventionToken: %s", subject.CSRFToken),
			"-H", fmt.Sprintf("Cookie: PVEAuthCookie=%s", subject.Ticket),
			"-H", "Content-Type: multipart/form-data",
			"--form", fmt.Sprintf("content=%s", ur.Format),
			"--form", fmt.Sprintf("filename=@%s", ur.File),
			uploadUrl(subject.ApiServer, ur.Node, ur.Storage))
		r, err := curl.CombinedOutput()
		if len(r) > 0 {
			output.Debug("\ncurl command output:\n\n {{index .output}}\n",
				map[string]interface{}{
					"event":  "curl_output",
					"output": string(r),
				})
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// Query the Proxmox API to match first storage device that accepts content specified by ProxmoxUploadRequest#Format
func (ur *ProxmoxUploadRequest) matchFirstStorageDevice() (string, error) {
	var (
		err     error
		subject *common.ProxmoxSubject
		req     *http.Request
		resp    *http.Response
		client  = common.HttpClient()
	)

	if subject, err = common.ReadSubjectFromCache(); err != nil {
		return "", fmt.Errorf("unable to read ticket cache: %s", err.Error())
	}

	if req, err = http.NewRequest(http.MethodGet, getStorageUrl(subject.ApiServer, ur.Node), nil); err != nil {
		return "", err
	} else if req, err = common.WithHttpCredentials(req); err != nil {
		return "", err
	}

	if resp, err = client.Do(req); err != nil {
		return "", err
	}
	defer resp.Body.Close()

	output.Debug("get storage request http code: {{index .code}}",
		map[string]interface{}{
			"event":  "http_response",
			"code":   resp.StatusCode,
			"status": resp.Status,
		})

	if resp.StatusCode == http.StatusUnauthorized {
		return "", errors.New("auth-error")
	} else if resp.StatusCode != http.StatusOK {
		return "", errors.New("get storage failed")
	}

	respData := make(map[string]interface{})
	if err = json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", shared.ErrParse
	}

	for _, each := range respData["data"].([]interface{}) {
		acceptContents := strings.Split(each.(map[string]interface{})["content"].(string), ",")
		for _, c := range acceptContents {
			if c == ur.Format {
				return each.(map[string]interface{})["storage"].(string), nil
			}
		}
	}

	return "", errors.New("no storage device match")
}

func uploadUrl(base, node, storage string) string {
	return fmt.Sprintf("%s/api2/json/nodes/%s/storage/%s/upload", base, node, storage)
}

func getStorageUrl(base, node string) string {
	return fmt.Sprintf("%s/api2/json/nodes/%s/storage", base, node)
}
