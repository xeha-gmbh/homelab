package login

import (
	"fmt"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

const (
	FlagUsername = "username"
	FlagPassword = "password"
	FlagApiServer = "api-server"

	DefaultUsername = "root@pam"
)

func NewProxmoxLoginCommand() *cobra.Command {
	payload := &ProxmoxLoginPayload{}

	cmd := &cobra.Command{
		Use: "login",
		Short: "login user with username and password",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.ParseFlags(args); err != nil {
				return nil
			}

			fmt.Println(payload.Username, payload.Password, payload.ApiServer)

			return nil
		},
	}

	addProxmoxLoginCommandFlags(cmd.PersistentFlags(), payload)
	markProxmoxLoginRequiredFlags(cmd)

	return cmd
}

func markProxmoxLoginRequiredFlags(cmd *cobra.Command) {
	for _, f := range []string{
		FlagPassword,
		FlagApiServer,
	} {
		cmd.MarkPersistentFlagRequired(f)
		cmd.MarkFlagRequired(f)
	}
}

// Bind proxmox login command flags to ProxmoxLoginPayload structure.
func addProxmoxLoginCommandFlags(flagSet *flag.FlagSet, payload *ProxmoxLoginPayload) {
	flagSet.StringVar(
		&payload.Username, FlagUsername, DefaultUsername,
		"The username that is authorized to carry out subsequent operations.",
		)
	flagSet.StringVar(
		&payload.Password, FlagPassword, "",
		"The password for the user. Required.",
		)
	flagSet.StringVar(
		&payload.ApiServer, FlagApiServer, "",
		"The address for the Proxmox API server. API paths will be appended to this address. Required.",
		)
}

type ProxmoxLoginPayload struct {
	Username	string
	Password 	string
	ApiServer 	string
}