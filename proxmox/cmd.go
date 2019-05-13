package proxmox

import (
	"github.com/xeha-gmbh/homelab/proxmox/login"
	"github.com/xeha-gmbh/homelab/proxmox/upload"
	"github.com/xeha-gmbh/homelab/proxmox/vm"
	"github.com/spf13/cobra"
)

func NewProxmoxCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxmox",
		Short: "easily interact with the proxmox platform for daily tasks",
	}

	cmd.AddCommand(login.NewProxmoxLoginCommand())
	cmd.AddCommand(upload.NewProxmoxUploadCommand())
	cmd.AddCommand(vm.NewProxmoxVMCommand())

	return cmd
}
