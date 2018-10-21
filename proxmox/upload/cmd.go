package upload

import (
	"github.com/imulab/homelab/proxmox/upload/api"
	"github.com/imulab/homelab/shared"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"os"
	"os/exec"
)

var (
	output shared.MessagePrinter
)

func NewProxmoxUploadCommand() *cobra.Command {
	payload := &ProxmoxUploadRequest{}

	cmd := &cobra.Command{
		Use:   "upload",
		Short: "upload file to Proxmox storage device",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.SetOutput(os.Stdout)

			if err := cmd.ParseFlags(args); err != nil {
				return err
			}

			output = shared.WithConfig(cmd, &payload.ExtraArgs)

			if err := checkCurlIsOnPath(); err != nil {
				output.Fatal(shared.ErrDependency.ExitCode,
					"Dependency unmet. Cause: {{index .cause}}",
					map[string]interface{}{
						"event": "pre_failed",
						"cause": err.Error(),
					})
				return shared.ErrDependency
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			err = payload.Upload()
			if err != nil {
				output.Fatal(shared.ErrOp.ExitCode,
					"Upload file {{index .file}} failed. Cause: {{index .cause}}",
					map[string]interface{}{
						"event": "upload_failed",
						"file":  payload.File,
						"cause": err.Error(),
					})
				return shared.ErrOp
			}

			output.Info("Upload file {{index .file}} is successful.",
				map[string]interface{}{
					"event": "upload_success",
					"file": payload.File,
				})
			return nil
		},
	}

	payload.InjectExtraArgs(cmd)
	addProxmoxLoginCommandFlags(cmd.PersistentFlags(), payload)
	markProxmoxUploadCommandRequiredFlags(cmd)

	return cmd
}

// Mark required upload command flags
func markProxmoxUploadCommandRequiredFlags(cmd *cobra.Command) {
	for _, f := range []string{
		api.FlagNode,
		api.FlagFile,
	} {
		cmd.MarkPersistentFlagRequired(f)
		cmd.MarkFlagRequired(f)
	}
}

// Bind proxmox upload command flags to ProxmoxUploadRequest structure.
func addProxmoxLoginCommandFlags(flagSet *flag.FlagSet, payload *ProxmoxUploadRequest) {
	flagSet.StringVar(
		&payload.Node, api.FlagNode, "",
		"The Proxmox cluster node that the upload operation targets. Required.",
	)
	flagSet.StringVar(
		&payload.Storage, api.FlagStorage, "",
		"The storage device label to upload file to. "+
			"If not set, command will query the node specified by --node to match the first storage device that accepts the file format --format.",
	)
	flagSet.StringVar(
		&payload.File, api.FlagFile, "",
		"The absolute path to the file to upload. Required.",
	)
	flagSet.StringVar(
		&payload.Format, api.FlagFormat, api.DefaultFormat,
		"The format of the file specified.",
	)
}

func checkCurlIsOnPath() error {
	if _, err := exec.LookPath("curl"); err != nil {
		return err
	}
	return nil
}
