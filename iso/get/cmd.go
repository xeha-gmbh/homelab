package get

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/xeha-gmbh/homelab/shared"
	"github.com/spf13/cobra"
)

const (
	flagFlavor    = "flavor"
	flagTargetDir = "target-dir"
	flagReuse     = "reuse"

	defaultTargetDir = "/tmp"
	defaultReuse     = false

	flavorUbuntuBionic64Live       = "ubuntu/bionic64.live"
	flavorUbuntuBionic64LiveUrl    = "http://releases.ubuntu.com/bionic/ubuntu-18.04.2-live-server-amd64.iso"
	flavorUbuntuBionic64NonLive    = "ubuntu/bionic64"
	flavorUbuntuBionic64NonLiveUrl = "http://cdimage.ubuntu.com/ubuntu/releases/18.04/release/ubuntu-18.04.2-server-amd64.iso"
	flavorUbuntuXenial64           = "ubuntu/xenial64"
	flavorUbuntuXenial64Url        = "http://releases.ubuntu.com/xenial/ubuntu-16.04.5-server-amd64.iso"

	noDefault = ""
)

type IsoGetPayload struct {
	ExtraArgs
	Flavor    string
	TargetDir string
	Reuse     bool
}

func NewIsoGetCommand() *cobra.Command {
	payload := new(IsoGetPayload)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "get system iso",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.SetOutput(os.Stdout)
			return cmd.ParseFlags(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				filename    string
				downloadUrl string
			)

			switch payload.Flavor {
			case flavorUbuntuBionic64Live:
				downloadUrl = flavorUbuntuBionic64LiveUrl
				filename = filepath.Join(payload.TargetDir, flavorUbuntuBionic64LiveUrl[strings.LastIndex(flavorUbuntuBionic64LiveUrl, "/")+1:])
			case flavorUbuntuBionic64NonLive:
				downloadUrl = flavorUbuntuBionic64NonLiveUrl
				filename = filepath.Join(payload.TargetDir, flavorUbuntuBionic64NonLiveUrl[strings.LastIndex(flavorUbuntuBionic64NonLiveUrl, "/")+1:])
			case flavorUbuntuXenial64:
				downloadUrl = flavorUbuntuXenial64Url
				filename = filepath.Join(payload.TargetDir, flavorUbuntuXenial64Url[strings.LastIndex(flavorUbuntuXenial64Url, "/")+1:])
			default:
				WithConfig(cmd, &payload.ExtraArgs).Fatal(
					1,
					"Flavor {{index .flavor}} is not supported.",
					map[string]interface{}{
						"event":     "unsupported_flavor",
						"flavor":    payload.Flavor,
						"exit-code": 1,
					})
				return errors.New("unsupported_flavor")
			}

			if _, err := os.Stat(filename); !os.IsNotExist(err) && payload.Reuse {
				WithConfig(cmd, &payload.ExtraArgs).Info(
					"Reused file at {{index .file}}, no download was executed.",
					map[string]interface{}{
						"event": "reused_file",
						"file":  filename,
						"reuse": payload.Reuse,
					})
				return nil
			}

			wgetArgs := []string{"-O", filename, downloadUrl}
			if !payload.Debug {
				wgetArgs = append([]string{"-q"}, wgetArgs...)
			}
			wget := exec.Command("wget", wgetArgs...)
			wget.Stdout = cmd.OutOrStdout()
			wget.Stderr = cmd.OutOrStderr()
			WithConfig(cmd, &payload.ExtraArgs).Debug(
				"Downloading from {{index .url}}, please wait.",
				map[string]interface{}{
					"event": "download_in_progress",
					"url":   downloadUrl,
				})
			if err := wget.Run(); err != nil {
				WithConfig(cmd, &payload.ExtraArgs).Fatal(
					2,
					"Download from {{index .url}} failed. Cause: {{index .cause}}",
					map[string]interface{}{
						"event":     "download_error",
						"url":       downloadUrl,
						"cause":     err.Error(),
						"exit-code": 2,
					})
				return errors.New("download_error")
			}

			WithConfig(cmd, &payload.ExtraArgs).Info(
				"Image {{index .flavor}} downloaded to {{index .file}}.",
				map[string]interface{}{
					"event":  "download_success",
					"flavor": payload.Flavor,
					"file":   filename,
				})
			return nil
		},
	}

	parseIsoGetCommandFlags(cmd, payload)
	markIsoGetCommandRequiredFlags(cmd)
	(&payload.ExtraArgs).InjectExtraArgs(cmd)

	return cmd
}

func markIsoGetCommandRequiredFlags(cmd *cobra.Command) {
	cmd.MarkFlagRequired(flagFlavor)
}

func parseIsoGetCommandFlags(cmd *cobra.Command, payload *IsoGetPayload) {
	cmd.Flags().StringVar(&payload.Flavor, flagFlavor, noDefault,
		"flavor of the image to download. ["+strings.Join([]string{
			flavorUbuntuBionic64Live,
			flavorUbuntuBionic64NonLive,
			flavorUbuntuXenial64,
		}, "|")+"]")
	cmd.Flags().StringVar(&payload.TargetDir, flagTargetDir, defaultTargetDir,
		"directory to put the downloaded put into.")
	cmd.Flags().BoolVar(&payload.Reuse, flagReuse, defaultReuse,
		"whether to use an existing image in the target directory if one is found.")
}
