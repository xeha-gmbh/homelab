package ticket

// Session information representing an authenticated Proxmox user
type Session struct {
	Username  string `json:"username"`
	Ticket    string `json:"ticket"`
	CSRFToken string `json:"csrf_token"`
	ApiServer string `json:"api_server"`
}
