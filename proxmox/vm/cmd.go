package vm

import (
	"github.com/imulab/homelab/proxmox/common"
	"github.com/spf13/cobra"
)

func NewProxmoxVMCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "vm",
		Short: "manage proxmox virtual machine",
	}

	cmd.AddCommand(NewProxmoxVMCreateCommand())

	return cmd
}

func NewProxmoxVMCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "create",
		Short: "create proxmox virtual machine",
	}

	for _, arch := range ArchetypeRepository().AllArchetypes() {
		subCmd := &cobra.Command{
			Use: arch.Use(),
			Short: arch.Short(),
			Long: arch.Long(),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := arch.CreateVM(); err != nil {
					return common.HandleError(err)
				}
				return nil
			},
		}

		arch.BindFlags(subCmd.Flags())
		for _, requiredFlag := range arch.RequiredFlags() {
			subCmd.MarkPersistentFlagRequired(requiredFlag)
			subCmd.MarkFlagRequired(requiredFlag)
		}

		cmd.AddCommand(subCmd)
	}

	return cmd
}

