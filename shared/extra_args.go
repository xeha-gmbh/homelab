package shared

import "github.com/spf13/cobra"

const (
	OutputFormatJson	= "json"
	OutputFormatText 	= "text"

	FlagDebug = "debug"
	FlagOutputFormat = "output-format"
)

type ExtraArgs struct {
	OutputFormat 	string
	Debug 			bool
}

func (ea *ExtraArgs) InjectExtraArgs(cmd *cobra.Command) {
	cmd.Flags().StringVar(&ea.OutputFormat, FlagOutputFormat, OutputFormatText,
		"Format of the output messages.")
	cmd.Flags().BoolVar(&ea.Debug, FlagDebug, false,
		"Whether to print debug messages.")
}