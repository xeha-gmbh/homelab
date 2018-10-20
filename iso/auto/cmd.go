package auto

import (
	"github.com/imulab/homelab/iso/auto/api"
	. "github.com/imulab/homelab/shared"
	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"os"
	"strings"
)

const (
	noDefault = ""
)

var (
	output MessagePrinter
)

type Payload struct {
	ExtraArgs
	Flavor      string `json:"flavor"`
	OutputPath  string `json:"output_path"`
	UsbBoot     bool   `json:"usb_boot"`
	Reuse       bool   `json:"reuse"`
	Timezone    string `json:"timezone"`
	Username    string `json:"username"`
	Password    string `json:"-"`
	Hostname    string `json:"hostname"`
	Domain      string `json:"domain"`
	IpAddress   string `json:"ip_address"`
	NetMask     string `json:"net_mask"`
	Gateway     string `json:"gateway"`
	NameServers string `json:"name_servers"`
}

func NewIsoAutoCommand() *cobra.Command {
	payload := new(Payload)

	cmd := &cobra.Command{
		Use:   "auto",
		Short: "create unattended installation media",
		Long: dedent.Dedent(`
			This command accepts an unmodified OS installation media and attempts to convert it
			into an unattended installation media by asking and answering all the installation
			questions in advance. The first supported OS is Ubuntu 18.04 LTS 64-bit, also known as
			ubuntu/bionic64. Future OS support will be added when needed.

			Thanks to https://github.com/netson/ubuntu-unattended for the wonderful script to pave
			the way. This command is largely based on the work of netson.
		`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.SetOutput(os.Stdout)
			if err := cmd.ParseFlags(args); err != nil {
				return err
			}
			output = WithConfig(cmd, &payload.ExtraArgs)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, provider := range []Provider{
				&UbuntuPreseedProvider{},
			} {
				if !provider.SupportsFlavor(payload.Flavor) {
					continue
				}

				if _, err := provider.CheckDependencies(payload); err != nil {
					output.Debug("Skipped provider {{index .providerName}} due to unmet dependency: {{index .cause}}",
						map[string]interface{}{
							"event":        "provider-skipped",
							"providerName": provider.Name(),
							"cause":        err.Error(),
						})
					continue
				}

				outputPath, err := provider.RemasterISO(payload)
				if err != nil {
					output.Fatal(ErrOp.ExitCode,
						"Provider {{index .providerName}} failed to remaster ISO. Cause: {{index .cause}}.",
						map[string]interface{}{
							"event":        "remaster-failed",
							"providerName": provider.Name(),
							"cause":        err.Error(),
						})
					return ErrOp
				}

				output.Info("Provider {{index .providerName}} successfully remastered ISO to {{index .outputPath}}.",
					map[string]interface{}{
						"event":        "remaster-success",
						"providerName": provider.Name(),
						"outputPath":   outputPath,
						"payload":      payload,
					})
				return nil
			}

			output.Fatal(ErrNoProvider.ExitCode,
				"No provider can handle flavor {{index .flavor}}",
				map[string]interface{}{
					"event":  "no-provider",
					"flavor": payload.Flavor,
				})
			return ErrNoProvider
		},
	}

	payload.InjectExtraArgs(cmd)
	addIsoAutoCommandFlags(cmd.Flags(), payload)
	markIsoAutoCommandRequiredFlags(cmd)

	return cmd
}

// Mark required auto command flags
func markIsoAutoCommandRequiredFlags(cmd *cobra.Command) {
	for _, f := range []string{
		api.FlagInputIso,
		api.FlagPassword,
		api.FlagHostname,
	} {
		cmd.MarkPersistentFlagRequired(f)
		cmd.MarkFlagRequired(f)
	}
}

// Bind 'iso auto' command flags to Payload structure.
func addIsoAutoCommandFlags(flagSet *flag.FlagSet, payload *Payload) {
	flagSet.StringVar(&payload.Flavor, api.FlagFlavor, api.DefaultFlavor,
		"An identification string for the OS. ["+strings.Join([]string{
			flavorUbuntuBionic64NonLive,
			flavorUbuntuXenial64,
		}, "|")+"]")
	flagSet.StringVar(&payload.OutputPath, api.FlagOutputPath, api.DefaultOutputPath,
		"Path where output files should be placed.")
	flagSet.BoolVar(&payload.UsbBoot, api.FlagUsbBoot, api.DefaultUsbBoot,
		"Whether the output ISO image should be made boot-able via USB.")
	flagSet.BoolVar(&payload.Reuse, api.FlagReuse, api.DefaultReuse,
		"Whether to reuse existing original images from the workspace.")
	flagSet.StringVar(&payload.Timezone, api.FlagTimezone, api.DefaultTimeZone,
		"Timezone of the new user.")
	flagSet.StringVar(&payload.Username, api.FlagUsername, api.DefaultUsername,
		"Username of the new user.")
	flagSet.StringVar(&payload.Password, api.FlagPassword, noDefault,
		"Password of the new user.")
	flagSet.StringVar(&payload.Hostname, api.FlagHostname, noDefault,
		"Hostname of the new system.")
	flagSet.StringVar(&payload.Domain, api.FlagDomain, api.DefaultDomain,
		"Domain of the new system.")
	flagSet.StringVar(&payload.IpAddress, api.FlagIpAddress, noDefault,
		"Ip address of the new system. Leave blank for DHCP auto configuration. "+
			"If set, should also set --net-mask, --gateway, and --name-servers")
	flagSet.StringVar(&payload.NetMask, api.FlagNetMask, api.DefaultNetMask,
		"Network mask of the specified network.")
	flagSet.StringVar(&payload.Gateway, api.FlagGateway, noDefault,
		"Network gateway of the specified network.")
	flagSet.StringVar(&payload.NameServers, api.FlagNameServers, api.DefaultNameServers,
		"A list of comma delimited DNS servers.")
}
