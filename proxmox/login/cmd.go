package login

import (
	"github.com/imulab/homelab/proxmox/login/api"
	"github.com/imulab/homelab/proxmox/login/impl"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

// Returns the 'login' command. This command expects to be installed
// as a sub-command where 'cmd.ParseFlags' has been called.
func NewProxmoxLoginCommand() *cobra.Command {
	service := impl.DefaultService()
	request := &api.Request{}

	cmd := &cobra.Command{
		Use:   "login",
		Short: "login user with username and password",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.ParseFlags(args); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := service.Login(request)
			if err != nil {
				return err
			}

			logrus.WithFields(logrus.Fields{
				"username": response.Username,
				"api":      response.ApiServer,
				"storage":  response.SessionStorage,
			}).Info("login successful.")

			return nil
		},
	}

	addProxmoxLoginCommandFlags(cmd.Flags(), request)
	markProxmoxLoginCommandRequiredFlags(cmd)

	return cmd
}

// Mark required login command flags
func markProxmoxLoginCommandRequiredFlags(cmd *cobra.Command) {
	for _, f := range []string{
		api.FlagPassword,
		api.FlagApiServer,
	} {
		cmd.MarkFlagRequired(f)
	}
}

// Bind proxmox login command flags to Request structure.
func addProxmoxLoginCommandFlags(flagSet *flag.FlagSet, request *api.Request) {
	flagSet.StringVar(
		&request.Username, api.FlagUsername, api.DefaultUsername,
		"The username that is authorized to carry out subsequent operations.",
	)
	flagSet.StringVar(
		&request.Password, api.FlagPassword, "",
		"The password for the user. Required.",
	)
	flagSet.StringVar(
		&request.Realm, api.FlagRealm, api.DefaultRealm,
		"The realm in Proxmox to log into.",
	)
	flagSet.StringVar(
		&request.ApiServer, api.FlagApiServer, "",
		"The address for the Proxmox API server. API paths will be appended to this address. Required.",
	)
	flagSet.BoolVar(
		&request.Force, api.FlagForce, api.DefaultForce,
		"If set, command will ignore existing ticket cache and force a re-login.",
	)
}
