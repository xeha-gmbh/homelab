package proxmox

import (
	"github.com/imulab/homelab/proxmox/login"
	"github.com/imulab/homelab/proxmox/upload"
	"github.com/imulab/homelab/proxmox/vm"
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
