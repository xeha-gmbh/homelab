package api

// Arguments for the 'proxmox login' command
type Request struct {
	Username  string
	Password  string
	Realm     string
	ApiServer string
	Force     bool
}
