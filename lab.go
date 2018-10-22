package main

import (
	"flag"
	"fmt"
	"github.com/imulab/homelab/bootstrap"
	"github.com/imulab/homelab/iso"
	"github.com/imulab/homelab/proxmox"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
	"strings"
)

func NewLabCommand() *cobra.Command {
	var (
		debug  bool
		format string
	)

	labCmd := &cobra.Command{
		Use:   "lab",
		Short: "lab: easily configure the lab environment",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.ParseFlags(args); err != nil {
				return err
			}

			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			} else {
				logrus.SetLevel(logrus.InfoLevel)
			}

			switch strings.ToLower(format) {
			case "json":
				logrus.SetFormatter(&logrus.JSONFormatter{})
			default:
				logrus.SetFormatter(&logrus.TextFormatter{})
			}

			return nil
		},
	}

	labCmd.ResetFlags()

	labCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "whether to print debug messages.")
	labCmd.PersistentFlags().StringVarP(&format, "output-format", "o", "text", "whether to print debug messages.")

	labCmd.AddCommand(proxmox.NewProxmoxCommand())
	labCmd.AddCommand(iso.NewIsoCommand())
	labCmd.AddCommand(bootstrap.NewBootstrapCommand())

	return labCmd
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
