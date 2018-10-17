package ubuntu

import (
	"fmt"
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

	autoScript = "ubuntu-auto.sh"
	autoScriptUrl = "https://raw.githubusercontent.com/imulab/homelab/iso/iso/auto/ubuntu/ubuntu-auto.sh"
)

type AutoIsoUbuntuProvider struct {}

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
	switch payload.Flavor {
	case flavorUbuntuBionic64:
		payload.Flavor = "bionic64"
	case flavorUbuntuXenial64:
		payload.Flavor = "xenial64"
	default:
		payload.Flavor = "unsupported"
	}

	if err := hashPassword(payload); err != nil {
		return err
	}

	parsedSeed, err := parseTemplateAndWriteToFile(payload)
	if err != nil {
		return err
	}

	remaster := exec.Command(
		filepath.Join(payload.OutputPath, autoScript),
		"--seed", parsedSeed,
		"--flavor", payload.Flavor,
		"--workspace", payload.OutputPath,
		"--reuse",
		"--bootable",
		"--debug")
	if out, err := remaster.CombinedOutput(); err != nil {
		fmt.Fprintln(os.Stdout, string(out))
		return NewGenericError(err.Error())
	} else {
		fmt.Fprintln(os.Stdout, string(out))
	}

	return nil
}

func hashPassword(payload *shared.Payload) error {
	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("echo %s | mkpasswd -s -m sha-512", payload.Password))

	if output, err := cmd.Output(); err != nil {
		return NewGenericError(err.Error())
	} else {
		payload.Password = string(output)
	}

	return nil
}

func parseTemplateAndWriteToFile(payload *shared.Payload) (parsedSeed string, err error) {
	var (
		f *os.File
		tmpl *template.Template
	)

	parsedSeed = filepath.Join(payload.OutputPath, "imulab.seed")

	if f, err = os.Create(parsedSeed); err != nil {
		return "", NewGenericError(err.Error())
	}
	if tmpl, err = template.New("seed").Parse(seedTmpl); err != nil {
		return "", NewGenericError(err.Error())
	}
	if err = tmpl.Execute(f, payload); err != nil {
		return "", NewGenericError(err.Error())
	}

	return
}
