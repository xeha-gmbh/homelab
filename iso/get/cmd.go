package get

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	flagFlavor 		= "flavor"
	flagTargetDir	= "target-dir"
	flagReuse 		= "reuse"

	defaultTargetDir 	= "/tmp"
	defaultReuse 		= false

	flavorUbuntuBionic64Live		= "ubuntu/bionic64.live"
	flavorUbuntuBionic64LiveUrl		= "http://releases.ubuntu.com/bionic/ubuntu-18.04.1-live-server-amd64.iso"
	flavorUbuntuBionic64NonLive		= "ubuntu/bionic64"
	flavorUbuntuBionic64NonLiveUrl	= "http://cdimage.ubuntu.com/ubuntu/releases/18.04/release/ubuntu-18.04.1-server-amd64.iso"
	flavorUbuntuXenial64			= "ubuntu/xenial64"
	flavorUbuntuXenial64Url 		= "http://releases.ubuntu.com/xenial/ubuntu-16.04.5-server-amd64.iso"

	noDefault	= ""
)

func NewIsoGetCommand() *cobra.Command {
	var (
		flavor 		string
		targetDir	string
		reuse 		bool
	)

	cmd := &cobra.Command{
		Use: "get",
		Short: "get system iso",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				filename string
				downloadUrl string
			)

			switch flavor {
			case flavorUbuntuBionic64Live:
				downloadUrl = flavorUbuntuBionic64LiveUrl
				filename = filepath.Join(targetDir, flavorUbuntuBionic64LiveUrl[strings.LastIndex(flavorUbuntuBionic64LiveUrl, "/")+1:])
			case flavorUbuntuBionic64NonLive:
				downloadUrl = flavorUbuntuBionic64NonLiveUrl
				filename = filepath.Join(targetDir, flavorUbuntuBionic64NonLiveUrl[strings.LastIndex(flavorUbuntuBionic64NonLiveUrl, "/")+1:])
			case flavorUbuntuXenial64:
				downloadUrl = flavorUbuntuXenial64Url
				filename = filepath.Join(targetDir, flavorUbuntuXenial64Url[strings.LastIndex(flavorUbuntuXenial64Url, "/")+1:])
			default:
				return errors.New("flavor not supported")
			}

			if _, err := os.Stat(filename); !os.IsNotExist(err) && reuse {
				fmt.Fprintf(os.Stdout, "Reusing file at %s, no download is executed.\n", filename)
				return nil
			}

			wget := exec.Command("wget", "-O", filename, downloadUrl)
			wget.Stdout = os.Stdout
			wget.Stderr = os.Stderr
			if err := wget.Run(); err != nil {
				return fmt.Errorf("download file %s to %s failed: %s", downloadUrl, filename, err.Error())
			}
			fmt.Fprintf(os.Stdout, "File downloaded to %s.\n", filename)

			return nil
		},
	}

	parseIsoGetCommandFlags(cmd, &flavor, &targetDir, &reuse)
	markIsoGetCommandRequiredFlags(cmd)

	return cmd
}

func markIsoGetCommandRequiredFlags(cmd *cobra.Command) {
	cmd.MarkFlagRequired(flagFlavor)
}

func parseIsoGetCommandFlags(cmd *cobra.Command, flavorAddr, targetDirAddr *string, reuseAddr *bool) {
	cmd.Flags().StringVar(flavorAddr, flagFlavor, noDefault,
		"flavor of the image to download. [" + strings.Join([]string{
			flavorUbuntuBionic64Live,

		}, "|") + "]")
	cmd.Flags().StringVar(targetDirAddr, flagTargetDir, defaultTargetDir,
		"directory to put the downloaded put into.")
	cmd.Flags().BoolVar(reuseAddr, flagReuse, defaultReuse,
		"whether to use an existing image in the target directory if one is found.")
}
