package ubuntu

import (
	"fmt"
	"github.com/imulab/homelab/iso/auto/shared"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	flavorUbuntuBionic64 = "ubuntu/bionic64"
	flavorUbuntuXenial64 = "ubuntu/xenial64"

	preseed 	= "preseed"
	seedName	= "imulab.seed"
)

type AutoIsoUbuntuProvider struct {}

func (p *AutoIsoUbuntuProvider) SupportsFlavor(flavor string) bool {
	fmt.Println("flavor", flavor)
	switch strings.ToLower(flavor) {
	case flavorUbuntuBionic64, flavorUbuntuXenial64:
		return true
	default:
		return false
	}
}

func (p *AutoIsoUbuntuProvider) CheckDependencies(payload *shared.Payload) (bool, error) {
	dependencies := map[string]string{
		"md5sum": "md5sum",
		"mkpasswd": "whois",
		"mkisofs": "genisoimage",
	}
	if payload.UsbBoot {
		dependencies["isohybrid"] = "[syslinux syslinux-utils]"
	}

	for dep, pkg := range dependencies {
		if _, err := exec.LookPath(dep); err != nil {
			return false, NewDependencyError(dep, pkg)
		}
	}

	return true, nil
}

func (p *AutoIsoUbuntuProvider) RemasterISO(payload *shared.Payload) error {
	var (
		err					error
		seedFilePath		string
		seedFileChecksum	string
	)

	// Clean up
	cleanUp := func() {
		for _, task := range []*exec.Cmd{
			unMountOriginalIso(payload),
			//removeDirectoryWithForce(originalDirectory(payload)),
			//removeDirectoryWithForce(remasterDirectory(payload)),
		} {
			task.Run()
		}
	}
	defer cleanUp()

	for _, dir := range []string{
		originalDirectory(payload),
		remasterDirectory(payload),
	} {
		if err = makeDirectory(dir); err != nil {
			return err
		}
	}

	for _, cmd := range []*exec.Cmd{
		mountOriginalIso(payload),
		copyDirectory(originalDirectory(payload), remasterDirectory(payload)),
		setInstallationMenuLanguage("en", payload),
		updateMenuTimeout(1, payload),
	} {
		if err = runCommand(cmd); err != nil {
			return err
		}
	}

	if err = hashPassword(payload); err != nil {
		return err
	}

	if seedFilePath, err = parseTemplateAndWriteToFile(payload); err != nil {
		return err
	}
	if seedFileChecksum, err = calculateChecksum(seedFilePath); err != nil {
		return err
	}

	if err = replaceAutoInstallOptions(seedFileChecksum, payload); err != nil {
		return err
	}

	fmt.Println("here")

	if err = createRemasterISO(payload); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Remastered auto install image saved to %s.",
		filepath.Join(remasterDirectory(payload), remasteredIsoName(payload)))
	return nil
}

func originalDirectory(payload *shared.Payload) string {
	return filepath.Join(payload.OutputPath, "iso-original")
}

func remasterDirectory(payload *shared.Payload) string {
	return filepath.Join(payload.OutputPath, "iso-remaster")
}

func remasteredIsoName(payload *shared.Payload) string {
	if len(payload.OutputName) == 0 {
		switch payload.Flavor {
		case flavorUbuntuXenial64:
			return "ubuntu-16.04-auto-install.iso"
		case flavorUbuntuBionic64:
			return "ubuntu-18.04-auto-install.iso"
		default:
			return "auto-install.iso"
		}
	}
	if strings.HasSuffix(strings.ToLower(payload.OutputName), ".iso") {
		return payload.OutputName
	} else {
		return payload.OutputName + ".iso"
	}
}

func mountOriginalIso(payload *shared.Payload) *exec.Cmd {
	return exec.Command("mount",
		"-o",
		"loop",
		payload.InputIso,
		originalDirectory(payload))
}

func unMountOriginalIso(payload *shared.Payload) *exec.Cmd {
	return exec.Command("umount", originalDirectory(payload))
}

func setInstallationMenuLanguage(lang string, payload *shared.Payload) *exec.Cmd {
	// Set language for installation menu, echo en > $tmp/iso_new/isolinux/lang
	return exec.Command("echo", lang, ">",
		filepath.Join(remasterDirectory(payload), "isolinux", "lang"))
}

func updateMenuTimeout(timeout int, payload *shared.Payload) *exec.Cmd {
	// sed -i -r 's/timeout\s+[0-9]+/timeout 1/g' $tmp/iso_new/isolinux/isolinux.cfg
	return exec.Command(
		"/bin/sh",
		"-c",
		fmt.Sprintf("sed -i -r 's/timeout\\s+[0-9]+/timeout %d/g' %s", timeout, filepath.Join(remasterDirectory(payload), "isolinux", "isolinux.cfg")))
}

func hashPassword(payload *shared.Payload) error {
	cmd := exec.Command("/bin/sh", "-c",
		fmt.Sprintf("echo %s | mkpasswd -s -m sha-512", payload.Password))

	if output, err := cmd.Output(); err != nil {
		return NewGenericError(err.Error())
	} else {
		fmt.Println(string(output))
		payload.Password = string(output)
	}

	return nil
}

func parseTemplateAndWriteToFile(payload *shared.Payload) (seedFilePath string, err error) {
	var (
		f *os.File
		tmpl *template.Template
	)

	seedFilePath = filepath.Join(remasterDirectory(payload), preseed, seedName)

	if f, err = os.Create(seedFilePath); err != nil {
		return "", NewGenericError(err.Error())
	}
	if tmpl, err = template.New("seed").Parse(seed); err != nil {
		return "", NewGenericError(err.Error())
	}
	if err = tmpl.Execute(f, payload); err != nil {
		return "", NewGenericError(err.Error())
	}

	return
}

func calculateChecksum(seedFilePath string) (string, error) {
	md5 := exec.Command("md5sum", seedFilePath)
	if sum, err := md5.Output(); err != nil {
		return "", NewGenericError(err.Error())
	} else {
		return strings.Split(string(sum), " ")[0], nil
	}
}

func replaceAutoInstallOptions(seedFileChecksum string, payload *shared.Payload) error {
	fmt.Println("checksum:", seedFileChecksum)
	sedCommand := fmt.Sprintf(`sed -i "/label install/ilabel autoinstall\n\menu label ^Autoinstall NETSON Ubuntu Server\n\kernel /install/vmlinuz\n\append file=/cdrom/preseed/ubuntu-server.seed initrd=/install/initrd.gz auto=true priority=high preseed/file=/cdrom/preseed/imulab.seed preseed/file/checksum=%s --" %s`, seedFileChecksum, filepath.Join(remasterDirectory(payload), "isolinux", "txt.cfg"))
	sed := exec.Command("/bin/sh", "-c", sedCommand)

	for _, arg := range sed.Args {
		fmt.Fprint(os.Stdout, arg, " ")
	}
	fmt.Println()

	if output, err := sed.CombinedOutput(); err != nil {
		fmt.Println(string(output))
		return NewGenericError(err.Error())
	}
	return nil
}

func createRemasterISO(payload *shared.Payload) error {
	base := remasterDirectory(payload)

	mk := exec.Command("mkisofs",
		"-D", "-r", "-V", "IMULAB_UBUNTU", "-cache-inodes", "-J", "-l",
		"-b", filepath.Join(base, "isolinux", "isolinux.bin"),
		"-c", filepath.Join(base, "isolinux", "boot.cat"),
		"-no-emul-boot", "-boot-load-size", "4", "-boot-info-table",
		"-o", filepath.Join(base, remasteredIsoName(payload)),
		base)

	if err := mk.Run(); err != nil {
		return NewGenericError(err.Error())
	}
	return nil
}