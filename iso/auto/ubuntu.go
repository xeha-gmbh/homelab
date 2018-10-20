package auto

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	flavorUbuntuBionic64NonLive = "ubuntu/bionic64"
	flavorUbuntuXenial64        = "ubuntu/xenial64"
	flavorBionic64              = "bionic64"
	flavorXenial64              = "xenial64"

	preseedScript    = "ubuntu-preseed.sh"
	preseedScriptUrl = "https://raw.githubusercontent.com/imulab/homelab/master/iso/auto/scripts/ubuntu-preseed.sh"

	preseedDefaultTemplate    = "preseed.default.tmpl"
	preseedDefaultTemplateUrl = "https://raw.githubusercontent.com/imulab/homelab/master/iso/auto/templates/preseed.default.tmpl"

	preseedName = "imulab.seed"

	flagSeed      = "--seed"
	flagFlavor    = "--flavor"
	flagInput     = "--input"
	flagOutput    = "--output"
	flagWorkspace = "--workspace"
	flagBootable  = "--bootable"
	flagReuse     = "--reuse"
	flagDebug     = "--debug"
)

type UbuntuPreseedProvider struct{}

func (p *UbuntuPreseedProvider) Name() string {
	return "ubuntu/preseed"
}

func (p *UbuntuPreseedProvider) SupportsFlavor(flavor string) bool {
	switch strings.ToLower(flavor) {
	case flavorUbuntuBionic64NonLive, flavorUbuntuXenial64:
		return true
	default:
		return false
	}
}

func (p *UbuntuPreseedProvider) CheckDependencies(payload *Payload) (bool, error) {
	if err := p.downloadPreseedScript(payload.Workspace); err != nil {
		return false, err
	} else if err := p.downloadDefaultPreseedTemplate(payload.Workspace); err != nil {
		return false, err
	}
	return true, nil
}

func (p *UbuntuPreseedProvider) RemasterISO(payload *Payload) (string, error) {
	// update flavor to adapt to the script '--flavor|-v' parameter
	switch payload.Flavor {
	case flavorUbuntuBionic64NonLive:
		payload.Flavor = flavorBionic64
	case flavorUbuntuXenial64:
		payload.Flavor = flavorXenial64
	default:
		payload.Flavor = "-"
	}

	// parse template
	parsedSeed, err := p.parseTemplateAndWriteToFile(payload)
	if err != nil {
		return "", err
	}

	// prepare arguments
	args := []string{
		flagSeed, parsedSeed,
		flagFlavor, payload.Flavor,
		flagWorkspace, payload.Workspace,
		flagInput, payload.InputIso,
		flagOutput, payload.OutputIso,
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
	remaster := exec.Command(filepath.Join(payload.Workspace, preseedScript), args...)
	remaster.Stdout = os.Stdout
	remaster.Stderr = os.Stderr
	if err := remaster.Start(); err != nil {
		return "", err
	}
	if err := remaster.Wait(); err != nil {
		return "", err
	}

	return payload.OutputIso, nil
}

func (p *UbuntuPreseedProvider) parseTemplateAndWriteToFile(payload *Payload) (string, error) {
	var (
		err        error
		sourcePath string
		targetPath string
		targetFile *os.File
		tmpl       *template.Template
	)

	sourcePath = filepath.Join(payload.Workspace, preseedDefaultTemplate)

	targetPath = filepath.Join(payload.Workspace, preseedName)
	if targetFile, err = os.Create(targetPath); err != nil {
		return "", err
	}

	if tmpl, err = template.New(filepath.Base(sourcePath)).ParseFiles(sourcePath); err != nil {
		return "", err
	}
	if err = tmpl.Execute(targetFile, payload); err != nil {
		return "", err
	}

	return targetPath, nil
}

func (p *UbuntuPreseedProvider) downloadDefaultPreseedTemplate(workspace string) error {
	if err := p.download(workspace, preseedDefaultTemplate, preseedDefaultTemplateUrl); err != nil {
		return err
	} else if err := os.Chmod(filepath.Join(workspace, preseedDefaultTemplate), 0644); err != nil {
		return err
	}
	return nil
}

func (p *UbuntuPreseedProvider) downloadPreseedScript(workspace string) error {
	if err := p.download(workspace, preseedScript, preseedScriptUrl); err != nil {
		return err
	} else if err := os.Chmod(filepath.Join(workspace, preseedScript), 0544); err != nil {
		return err
	}
	return nil
}

func (p *UbuntuPreseedProvider) download(workspace, filename, url string) error {
	var (
		err  error
		path = filepath.Join(workspace, filename)
		f    *os.File
		resp *http.Response
	)

	if f, err = os.Create(path); err != nil {
		return err
	}
	defer f.Close()

	if resp, err = http.Get(url); err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download error from %s: code %d", url, resp.StatusCode)
	}
	defer resp.Body.Close()

	if _, err = io.Copy(f, resp.Body); err != nil {
		return err
	}

	return nil
}
