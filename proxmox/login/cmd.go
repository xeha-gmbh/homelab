package login

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

const (
	FlagUsername  = "username"
	FlagPassword  = "password"
	FlagRealm     = "realm"
	FlagApiServer = "api-server"
	FlagForce     = "force"

	DefaultUsername = "root"
	DefaultRealm    = "pam"
	DefaultForce    = false

	TicketCache = ".proxmox"
)

// Returns the 'login' command. This command expects to be installed
// as a sub-command where 'cmd.ParseFlags' has been called.
func NewProxmoxLoginCommand() *cobra.Command {
	payload := &ProxmoxLoginRequest{}

	cmd := &cobra.Command{
		Use:   "login",
		Short: "login user with username and password",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err          error
				subject      *ProxmoxSubject
				isNewAttempt bool
			)

			subject, isNewAttempt, err = payload.Login()
			if err != nil {
				return handleError(err)
			}

			if isNewAttempt {
				err = subject.WriteToFile(proxmoxTicketCache())
				if err != nil {
					return handleError(err)
				}
			}

			return nil
		},
	}

	addProxmoxLoginCommandFlags(cmd.PersistentFlags(), payload)
	markProxmoxLoginCommandRequiredFlags(cmd)

	return cmd
}

// Mark required login command flags
func markProxmoxLoginCommandRequiredFlags(cmd *cobra.Command) {
	for _, f := range []string{
		FlagPassword,
		FlagApiServer,
	} {
		cmd.MarkPersistentFlagRequired(f)
		cmd.MarkFlagRequired(f)
	}
}

// Bind proxmox login command flags to ProxmoxLoginRequest structure.
func addProxmoxLoginCommandFlags(flagSet *flag.FlagSet, payload *ProxmoxLoginRequest) {
	flagSet.StringVar(
		&payload.Username, FlagUsername, DefaultUsername,
		"The username that is authorized to carry out subsequent operations.",
	)
	flagSet.StringVar(
		&payload.Password, FlagPassword, "",
		"The password for the user. Required.",
	)
	flagSet.StringVar(
		&payload.Realm, FlagRealm, DefaultRealm,
		"The realm in Proxmox to log into.",
	)
	flagSet.StringVar(
		&payload.ApiServer, FlagApiServer, "",
		"The address for the Proxmox API server. API paths will be appended to this address. Required.",
	)
	flagSet.BoolVar(
		&payload.Force, FlagForce, DefaultForce,
		"If set, command will ignore existing ticket cache and force a re-login.",
	)
}
