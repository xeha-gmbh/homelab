package upload

import (
	"github.com/imulab/homelab/proxmox/common"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

const (
	FlagNode = "node"
	FlagStorage = "storage"
	FlagFile = "file"
	FlagFormat = "format"

	DefaultFormat = "iso"
)

func NewProxmoxUploadCommand() *cobra.Command {
	payload := &ProxmoxUploadRequest{}

	cmd := &cobra.Command{
		Use:   "upload",
		Short: "upload file to Proxmox storage device",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
			)

			err = payload.Upload()
			if err != nil {
				return common.HandleError(err)
			}

			return nil
		},
	}

	addProxmoxLoginCommandFlags(cmd.PersistentFlags(), payload)
	markProxmoxUploadCommandRequiredFlags(cmd)

	return cmd
}


// Mark required upload command flags
func markProxmoxUploadCommandRequiredFlags(cmd *cobra.Command) {
	for _, f := range []string{
		FlagNode,
		FlagFile,
	} {
		cmd.MarkPersistentFlagRequired(f)
		cmd.MarkFlagRequired(f)
	}
}

// Bind proxmox upload command flags to ProxmoxUploadRequest structure.
func addProxmoxLoginCommandFlags(flagSet *flag.FlagSet, payload *ProxmoxUploadRequest) {
	flagSet.StringVar(
		&payload.Node, FlagNode, "",
		"The Proxmox cluster node that the upload operation targets. Required.",
	)
	flagSet.StringVar(
		&payload.Storage, FlagStorage, "",
		"The storage device label to upload file to. " +
			"If not set, command will query the node specified by --node to match the first storage device that accepts the file format --format.",
	)
	flagSet.StringVar(
		&payload.File, FlagFile, "",
		"The absolute path to the file to upload. Required.",
	)
	flagSet.StringVar(
		&payload.Format, FlagFormat, DefaultFormat,
		"The format of the file specified.",
	)
}