package bootstrap

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	flagConfig = "config"
	noDefault = ""
)

func NewBootstrapCommand() *cobra.Command {
	var yamlPath string

	cmd := &cobra.Command{
		Use:"bootstrap",
		Short:"bootstrap home lab with a single config",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := ParseConfig(yamlPath)
			if err != nil {
				return err
			}
			return config.Bootstrap()
		},
	}

	pflag.StringVar(&yamlPath, flagConfig, noDefault, "Path to the YAML configuration file.")
	cmd.MarkFlagFilename(flagConfig, "yaml", "yml")
	cmd.MarkFlagRequired(flagConfig)

	return cmd
}
