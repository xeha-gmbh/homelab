package iso

import (
	"github.com/xeha-gmbh/homelab/iso/auto"
	"github.com/xeha-gmbh/homelab/iso/get"
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
