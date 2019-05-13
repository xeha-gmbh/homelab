package login

import (
	"github.com/xeha-gmbh/homelab/proxmox/common"
	"github.com/xeha-gmbh/homelab/proxmox/login/api"
	. "github.com/xeha-gmbh/homelab/shared"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"os"
)

var (
	output MessagePrinter
)

// Returns the 'login' command. This command expects to be installed
// as a sub-command where 'cmd.ParseFlags' has been called.
func NewProxmoxLoginCommand() *cobra.Command {
	payload := &ProxmoxLoginRequest{}

	cmd := &cobra.Command{
		Use:   "login",
		Short: "login user with username and password",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.SetOutput(os.Stdout)
			if err := cmd.ParseFlags(args); err != nil {
				return err
			}
			output = WithConfig(cmd, &payload.ExtraArgs)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err          error
				subject      *common.ProxmoxSubject
				isNewAttempt bool
			)

			subject, isNewAttempt, err = payload.Login()
			if err != nil {
				output.Fatal(ErrOp.ExitCode,
					"Failed to login. Cause: {{index .cause}}",
					map[string]interface{}{
						"event": "login_failed",
						"cause": err.Error(),
					})
				return ErrOp
			}

			if isNewAttempt {
				err = common.WriteSubjectToCache(subject)
				if err != nil {
					output.Fatal(ErrOp.ExitCode,
						"Failed to save cache. Cause: {{index .cause}}",
						map[string]interface{}{
							"event": "cache_save_failed",
							"cause": err.Error(),
						})
					return ErrOp
				}
			}

			return nil
		},
	}

	payload.InjectExtraArgs(cmd)
	addProxmoxLoginCommandFlags(cmd.PersistentFlags(), payload)
	markProxmoxLoginCommandRequiredFlags(cmd)

	return cmd
}

// Mark required login command flags
func markProxmoxLoginCommandRequiredFlags(cmd *cobra.Command) {
	for _, f := range []string{
		api.FlagPassword,
		api.FlagApiServer,
	} {
		cmd.MarkPersistentFlagRequired(f)
		cmd.MarkFlagRequired(f)
	}
}

// Bind proxmox login command flags to ProxmoxLoginRequest structure.
func addProxmoxLoginCommandFlags(flagSet *flag.FlagSet, payload *ProxmoxLoginRequest) {
	flagSet.StringVar(
		&payload.Username, api.FlagUsername, api.DefaultUsername,
		"The username that is authorized to carry out subsequent operations.",
	)
	flagSet.StringVar(
		&payload.Password, api.FlagPassword, "",
		"The password for the user. Required.",
	)
	flagSet.StringVar(
		&payload.Realm, api.FlagRealm, api.DefaultRealm,
		"The realm in Proxmox to log into.",
	)
	flagSet.StringVar(
		&payload.ApiServer, api.FlagApiServer, "",
		"The address for the Proxmox API server. API paths will be appended to this address. Required.",
	)
	flagSet.BoolVar(
		&payload.Force, api.FlagForce, api.DefaultForce,
		"If set, command will ignore existing ticket cache and force a re-login.",
	)
}
