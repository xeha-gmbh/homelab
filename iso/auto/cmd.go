package auto

import (
	"errors"
	"fmt"
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

var (
	allProviders = make([]Provider, 0)
)

type Payload struct {
	// flavor of the OS
	Flavor 		string
	// path to the input ISO
	InputIso 	string
	// path of the output files
	OutputPath 	string
	// name of the output ISO file
	OutputName 	string
	// if should be made bootable via USB
	UsbBoot		bool

	// Attributes of the new user
	Timezone 	string
	Username 	string
	Password 	string
	Hostname 	string
	Domain 		string
}

func NewIsoAutoCommand() *cobra.Command {
	payload := &Payload{}

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
			for _, provider := range allProviders {
				if provider.SupportsFlavor(payload.Flavor) {
					if _, err := provider.CheckDependencies(); err != nil {
						fmt.Fprintf(os.Stdout, "provider skipped due to unmet dependency: %s\n", err.Error())
						continue
					}
					return handleError(provider.RemasterISO(payload))
				}
			}
			return handleError(errors.New(errorNoProvider))
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
func addProxmoxLoginCommandFlags(flagSet *flag.FlagSet, payload *Payload) {
	flagSet.StringVar(&payload.Flavor, FlagFlavor, DefaultFlavor,
		"An identification string for the OS. [ubuntu/bionic64 | ubuntu/xenial64]")
	flagSet.StringVar(&payload.InputIso, FlagInputIso, noDefault,
		"Path to the original ISO image.")
	flagSet.StringVar(&payload.OutputPath, FlagOutputPath, DefaultOutputPath,
		"Path where output files should be placed.")
	flagSet.StringVar(&payload.OutputName, FlagOutputName, noDefault,
		"Name of the output remastered ISO image. Determined by provider if unset.")
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
