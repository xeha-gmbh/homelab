package proxmox

import (
	"github.com/imulab/homelab/proxmox/login"
	"github.com/spf13/cobra"
)

func NewProxmoxCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxmox",
		Short: "easily start a set of vms on proxmox",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(login.NewProxmoxLoginCommand())

	return cmd
}
