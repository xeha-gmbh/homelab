package api

// Success response for a 'proxmox login' request.
// (Error responses are delivered via error)
type Response struct {
	Username       string `json:"username"`
	ApiServer      string `json:"api_server"`
	SessionStorage string `json:"session_storage"`
}
