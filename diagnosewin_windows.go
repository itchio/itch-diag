// +build windows

package main

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows/registry"
)

func (a *App) DiagnoseWindows() {
	a.Debugf("Running Windows diagnostics...")
	a.Test("Collecting system information with WMIC", a.DiagnoseWmic)
	a.Test("Diagnosing NUL registry hack", a.DiagnoseNUL)
	a.Test("Verifying installed app information", a.DiagnoseItchReg)
}

func (a *App) DiagnoseNUL() error {
	// TODO: implement
	return nil
}

func (a *App) DiagnoseWmic() error {
	out, err := exec.Command("wmic", "os", "get",
		"Caption,",
		"Version,",
		"BuildNumber,",
		"OSArchitecture,",
		"InstallDate,",
		"WindowsDirectory,",
		"MUILanguages",
		"/format:list",
	).CombinedOutput()
	if err != nil {
		return errors.WithStack(err)
	}

	result := string(out)
	result = strings.Replace(result, "\r\n", "\n", -1)
	result = strings.TrimSpace(result)
	a.Infof("OS information:\n<pre>%s</pre>", result)

	return nil
}

const uninstallRegPrefix = "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall"

func (a *App) DiagnoseItchReg() error {
	pk, err := registry.OpenKey(registry.CURRENT_USER, uninstallRegPrefix, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return errors.WithStack(err)
	}
	defer pk.Close()

	k, err := registry.OpenKey(pk, "itch", registry.READ)
	if err != nil {
		return errors.WithStack(err)
	}
	defer k.Close()

	installFolder, _, err := k.GetStringValue("InstallLocation")
	if err != nil {
		return errors.WithStack(err)
	}

	err = a.DiagnoseInstallFolder(installFolder)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
