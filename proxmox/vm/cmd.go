package vm

import (
	"github.com/xeha-gmbh/homelab/shared"
	"github.com/spf13/cobra"
	"os"
)

var (
	output shared.MessagePrinter
)

func NewProxmoxVMCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vm",
		Short: "manage proxmox virtual machine",
	}

	cmd.AddCommand(NewProxmoxVMCreateCommand())

	return cmd
}

func NewProxmoxVMCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create proxmox virtual machine",
	}

	for _, arch := range ArchetypeRepository().AllArchetypes() {
		subCmd := &cobra.Command{
			Use:   arch.Use(),
			Short: arch.Short(),
			Long:  arch.Long(),
			PreRunE: func(cmd *cobra.Command, args []string) error {
				cmd.SetOutput(os.Stdout)
				if err := cmd.ParseFlags(args); err != nil {
					return err
				}
				return nil
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := arch.CreateVM(); err != nil {
					return err
				}
				return nil
			},
		}

		arch.BindFlags(subCmd)
		for _, requiredFlag := range arch.RequiredFlags() {
			subCmd.MarkPersistentFlagRequired(requiredFlag)
			subCmd.MarkFlagRequired(requiredFlag)
		}

		cmd.AddCommand(subCmd)
	}

	return cmd
}
