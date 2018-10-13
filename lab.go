package main

import (
	"flag"
	"fmt"
	"github.com/imulab/homelab/proxmox"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
)

func NewLabCommand() *cobra.Command {
	cmds := &cobra.Command{
		Use:   "lab",
		Short: "lab: easily configure the lab environment",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.ParseFlags(args); err != nil {
				return err
			}
			return nil
		},
	}

	cmds.ResetFlags()
	cmds.AddCommand(proxmox.NewProxmoxCommand())

	return cmds
}

func Run() error {
	// TODO pflag.CommandLine.SetNormalizeFunc()

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	cmd := NewLabCommand()
	return cmd.Execute()
}

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
