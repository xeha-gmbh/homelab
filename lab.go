package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/xeha-gmbh/homelab/bootstrap"
	"github.com/xeha-gmbh/homelab/iso"
	"github.com/xeha-gmbh/homelab/proxmox"
)

var (
	once   sync.Once
	labCmd *cobra.Command
)

func GetLabCommand() *cobra.Command {
	once.Do(func() {
		labCmd = &cobra.Command{
			Use:   "lab",
			Short: "lab: easily configure the lab environment",
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				if err := cmd.ParseFlags(args); err != nil {
					return err
				}
				return nil
			},
		}

		labCmd.ResetFlags()
		labCmd.AddCommand(proxmox.NewProxmoxCommand())
		labCmd.AddCommand(iso.NewIsoCommand())
		labCmd.AddCommand(bootstrap.NewBootstrapCommand())
	})

	return labCmd
}

func Run() error {
	// TODO pflag.CommandLine.SetNormalizeFunc()

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	cmd := GetLabCommand()
	return cmd.Execute()
}

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
