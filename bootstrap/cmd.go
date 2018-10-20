package bootstrap

import (
	. "github.com/imulab/homelab/shared"
	"github.com/spf13/cobra"
)

const (
	flagConfig = "config"
	noDefault  = ""
)

var (
	output    MessagePrinter
	extraArgs *ExtraArgs
)

func NewBootstrapCommand() *cobra.Command {
	payload := new(Payload)

	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "bootstrap home lab with a single config",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.ParseFlags(args); err != nil {
				return err
			}
			extraArgs = &payload.ExtraArgs
			output = WithConfig(cmd, extraArgs)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := ParseConfig(payload.YamlPath)
			if err != nil {
				return err
			}
			return config.Bootstrap()
		},
	}

	cmd.Flags().StringVar(&payload.YamlPath, flagConfig, noDefault, "Path to the YAML configuration file.")
	cmd.MarkFlagFilename(flagConfig, "yaml", "yml")
	cmd.MarkFlagRequired(flagConfig)
	payload.ExtraArgs.InjectExtraArgs(cmd)

	return cmd
}

type Payload struct {
	ExtraArgs
	YamlPath string
}
