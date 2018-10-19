package iso

import (
	"github.com/imulab/homelab/iso/auto"
	"github.com/imulab/homelab/iso/get"
	"github.com/spf13/cobra"
)

func NewIsoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iso",
		Short: "utility to enhance iso images",
	}

	cmd.AddCommand(auto.NewIsoAutoCommand())
	cmd.AddCommand(get.NewIsoGetCommand())

	return cmd
}
