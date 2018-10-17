package auto

import (
	"fmt"
	"github.com/imulab/homelab/iso/auto/shared"
	"github.com/imulab/homelab/iso/auto/ubuntu"
	"github.com/imulab/homelab/iso/common"
	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"os"
)

const (
	FlagFlavor      = "flavor"
	FlagInputIso    = "iso"
	FlagOutputPath  = "target-dir"
	FlagUsbBoot     = "usb-boot"
	FlagDebug       = "debug"
	FlagReuse       = "reuse"
	FlagTimezone    = "timezone"
	FlagUsername    = "username"
	FlagPassword    = "password"
	FlagHostname    = "hostname"
	FlagDomain      = "domain"
	FlagIpAddress   = "ip-address"
	FlagNetMask     = "net-mask"
	FlagGateway     = "gateway"
	FlagNameServers = "name-servers"

	DefaultFlavor      = "ubuntu/bionic64"
	DefaultOutputPath  = "/tmp"
	DefaultUsbBoot     = true
	DefaultDebug       = false
	DefaultReuse       = false
	DefaultTimeZone    = "America/Toronto"
	DefaultUsername    = "imulab"
	DefaultDomain      = "home.local"
	DefaultNetMask     = "255.255.255.0"
	DefaultNameServers = "8.8.8.8"

	noDefault = ""
)

func NewIsoAutoCommand() *cobra.Command {
	payload := &shared.Payload{}

	cmd := &cobra.Command{
		Use:   "auto",
		Short: "create unattended installation media",
		Long: dedent.Dedent(`
			This command accepts an unmodified OS installation media and attempts to convert it
			into an unattended installation media by asking and answering all the installation
			questions in advance. The first supported OS is Ubuntu 18.04 LTS 64-bit, also known as
			ubuntu/bionic64. Future OS support will be added when needed.

			Thanks to https://github.com/netson/ubuntu-unattended for the wonderful script to pave
			the way. This command is largely based on the work of neston.
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, provider := range []shared.Provider{
				&ubuntu.AutoIsoUbuntuProvider{},
			} {
				if provider.SupportsFlavor(payload.Flavor) {
					if _, err := provider.CheckDependencies(payload); err != nil {
						fmt.Fprintf(os.Stdout, "provider skipped due to unmet dependency: %s\n", err.Error())
						continue
					}
					return common.HandleError(provider.RemasterISO(payload))
				}
			}
			return common.HandleError(shared.NoProviderError{})
		},
	}

	addIsoAutoCommandFlags(cmd.Flags(), payload)
	markIsoAutoCommandRequiredFlags(cmd)

	return cmd
}

// Mark required auto command flags
func markIsoAutoCommandRequiredFlags(cmd *cobra.Command) {
	for _, f := range []string{
		FlagInputIso,
		FlagPassword,
		FlagHostname,
	} {
		cmd.MarkPersistentFlagRequired(f)
		cmd.MarkFlagRequired(f)
	}
}

// Bind 'iso auto' command flags to Payload structure.
func addIsoAutoCommandFlags(flagSet *flag.FlagSet, payload *shared.Payload) {
	flagSet.StringVar(&payload.Flavor, FlagFlavor, DefaultFlavor,
		"An identification string for the OS. [ubuntu/bionic64 | ubuntu/xenial64]")
	flagSet.StringVar(&payload.OutputPath, FlagOutputPath, DefaultOutputPath,
		"Path where output files should be placed.")
	flagSet.BoolVar(&payload.UsbBoot, FlagUsbBoot, DefaultUsbBoot,
		"Whether the output ISO image should be made boot-able via USB.")
	flagSet.BoolVar(&payload.Debug, FlagDebug, DefaultDebug,
		"Whether to print debug message during execution.")
	flagSet.BoolVar(&payload.Reuse, FlagReuse, DefaultReuse,
		"Whether to reuse existing original images from the workspace.")
	flagSet.StringVar(&payload.Timezone, FlagTimezone, DefaultTimeZone,
		"Timezone of the new user.")
	flagSet.StringVar(&payload.Username, FlagUsername, DefaultUsername,
		"Username of the new user.")
	flagSet.StringVar(&payload.Password, FlagPassword, noDefault,
		"Password of the new user.")
	flagSet.StringVar(&payload.Hostname, FlagHostname, noDefault,
		"Hostname of the new system.")
	flagSet.StringVar(&payload.Domain, FlagDomain, DefaultDomain,
		"Domain of the new system.")
	flagSet.StringVar(&payload.IpAddress, FlagIpAddress, noDefault,
		"Ip address of the new system. Leave blank for DHCP auto configuration. "+
			"If set, should also set --net-mask, --gateway, and --name-servers")
	flagSet.StringVar(&payload.NetMask, FlagNetMask, DefaultNetMask,
		"Network mask of the specified network.")
	flagSet.StringVar(&payload.Gateway, FlagGateway, noDefault,
		"Network gateway of the specified network.")
	flagSet.StringVar(&payload.NameServers, FlagNameServers, DefaultNameServers,
		"A list of comma delimited DNS servers.")
}
