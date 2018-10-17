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
	FlagFlavor 		= "flavor"
	FlagInputIso	= "iso"
	FlagOutputPath	= "target-dir"
	FlagOutputName 	= "target-name"
	FlagUsbBoot 	= "usb-boot"
	FlagTimezone 	= "timezone"
	FlagUsername 	= "username"
	FlagPassword	= "password"
	FlagHostname 	= "hostname"
	FlagDomain 		= "domain"

	DefaultFlavor			= "ubuntu/bionic64"
	DefaultOutputPath		= "/tmp"
	DefaultOutputNameTmpl	= "%s-unattended.iso"
	DefaultUsbBoot 			= true
	DefaultTimeZone 		= "America/Toronto"
	DefaultUsername 		= "imulab"
	DefaultDomain 			= "home.local"

	noDefault = ""
)

func NewIsoAutoCommand() *cobra.Command {
	payload := &shared.Payload{}

	cmd := &cobra.Command{
		Use: "auto",
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

	addProxmoxLoginCommandFlags(cmd.Flags(), payload)
	markProxmoxLoginCommandRequiredFlags(cmd)

	return cmd
}

// Mark required auto command flags
func markProxmoxLoginCommandRequiredFlags(cmd *cobra.Command) {
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
func addProxmoxLoginCommandFlags(flagSet *flag.FlagSet, payload *shared.Payload) {
	flagSet.StringVar(&payload.Flavor, FlagFlavor, DefaultFlavor,
		"An identification string for the OS. [ubuntu/bionic64 | ubuntu/xenial64]")
	flagSet.StringVar(&payload.OutputPath, FlagOutputPath, DefaultOutputPath,
		"Path where output files should be placed.")
	flagSet.BoolVar(&payload.UsbBoot, FlagUsbBoot, DefaultUsbBoot,
		"Whether the output ISO image should be made boot-able via USB.")
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
}
