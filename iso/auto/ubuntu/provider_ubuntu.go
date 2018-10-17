package ubuntu

import (
	"github.com/imulab/homelab/iso/auto/shared"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	flavorUbuntuBionic64 = "ubuntu/bionic64"
	flavorUbuntuXenial64 = "ubuntu/xenial64"
	flavorBionic64       = "bionic64"
	flavorXenial64       = "xenial64"

	autoScript    = "ubuntu-auto.sh"
	autoScriptUrl = "https://raw.githubusercontent.com/imulab/homelab/iso/iso/auto/ubuntu/ubuntu-auto.sh"

	flagSeed      = "--seed"
	flagFlavor    = "--flavor"
	flagWorkspace = "--workspace"
	flagBootable  = "--bootable"
	flagReuse     = "--reuse"
	flagDebug     = "--debug"
)

type AutoIsoUbuntuProvider struct{}

func (p *AutoIsoUbuntuProvider) SupportsFlavor(flavor string) bool {
	switch strings.ToLower(flavor) {
	case flavorUbuntuBionic64, flavorUbuntuXenial64:
		return true
	default:
		return false
	}
}

func (p *AutoIsoUbuntuProvider) CheckDependencies(payload *shared.Payload) (bool, error) {
	if err := downloadAutoScript(payload.OutputPath); err != nil {
		return false, err
	}
	return true, nil
}

func downloadAutoScript(workspace string) error {
	scriptPath := filepath.Join(workspace, autoScript)

	out, err := os.Create(scriptPath)
	if err != nil {
		return NewGenericError(err.Error())
	}
	defer out.Close()

	resp, err := http.Get(autoScriptUrl)
	if err != nil {
		return NewGenericError(err.Error())
	} else if resp.StatusCode != http.StatusOK {
		return NewGenericError("failed to download auto script")
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return NewGenericError(err.Error())
	}

	err = os.Chmod(scriptPath, 0544)
	if err != nil {
		return NewGenericError(err.Error())
	}

	return nil
}

func (p *AutoIsoUbuntuProvider) RemasterISO(payload *shared.Payload) error {
	// update flavor to adapt to the script '--flavor|-v' parameter
	switch payload.Flavor {
	case flavorUbuntuBionic64:
		payload.Flavor = flavorBionic64
	case flavorUbuntuXenial64:
		payload.Flavor = flavorXenial64
	default:
		payload.Flavor = "-"
	}

	// parse template
	parsedSeed, err := parseTemplateAndWriteToFile(payload)
	if err != nil {
		return err
	}

	// prepare arguments
	args := []string{
		flagSeed, parsedSeed,
		flagFlavor, payload.Flavor,
		flagWorkspace, payload.OutputPath,
	}
	if payload.UsbBoot {
		args = append(args, flagBootable)
	}
	if payload.Debug {
		args = append(args, flagDebug)
	}
	if payload.Reuse {
		args = append(args, flagReuse)
	}

	// execute command
	remaster := exec.Command(filepath.Join(payload.OutputPath, autoScript), args...)
	remaster.Stdout = os.Stdout
	remaster.Stderr = os.Stderr
	if err := remaster.Start(); err != nil {
		return NewGenericError(err.Error())
	}
	if err := remaster.Wait(); err != nil {
		return NewGenericError(err.Error())
	}

	return nil
}

func parseTemplateAndWriteToFile(payload *shared.Payload) (parsedSeed string, err error) {
	var (
		f    *os.File
		tmpl *template.Template
	)

	parsedSeed = filepath.Join(payload.OutputPath, seedName)

	if f, err = os.Create(parsedSeed); err != nil {
		return "", NewGenericError(err.Error())
	}
	if tmpl, err = template.New(seedName).Parse(seedTemplate); err != nil {
		return "", NewGenericError(err.Error())
	}
	if err = tmpl.Execute(f, payload); err != nil {
		return "", NewGenericError(err.Error())
	}

	return
}
