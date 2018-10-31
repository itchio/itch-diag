// +build windows

package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows/registry"
)

func (a *App) DiagnoseWindows() {
	a.Debugf("Running Windows diagnostics...")
	a.Test("Collecting OS information", a.CollectOSInfo)
	a.Test("Collecting Security Center information", a.CollectSecurityInfo)
	a.Test("Verifying null service", a.DiagnoseNUL)
	a.Test("Verifying installed app information", a.DiagnoseItchReg)
}

const nullServiceRegPath = "SYSTEM\\ControlSet001\\Services\\Null"

var nullServiceFullPath = fmt.Sprintf("HKLM\\%s\\Start", nullServiceRegPath)

func (a *App) DiagnoseNUL() error {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, nullServiceRegPath, registry.READ)
	if err != nil {
		return errors.WithStack(err)
	}
	defer k.Close()

	startValue, _, err := k.GetIntegerValue("Start")
	if err != nil {
		return errors.WithStack(err)
	}

	if startValue != 1 {
		a.Errorf("Null service has been tinkered with!")
		a.Errorf("<code>%s</code> should be 1, but is currently %d", nullServiceRegPath, startValue)
	}

	return nil
}

func (a *App) CollectOSInfo() error {
	res, err := a.RunWmic(WmicOptions{
		Alias: "os",
		Columns: []string{
			"Caption",
			"Version",
			"BuildNumber",
			"OSArchitecture",
			"InstallDate",
			"WindowsDirectory",
			"MUILanguages",
		},
	})
	if err != nil {
		return errors.WithStack(err)
	}

	a.Infof("%s (%s), version %s, build %s",
		res["Caption"],
		res["OSArchitecture"],
		res["Version"],
		res["BuildNumber"],
	)
	a.Infof("Installed %s", res["InstallDate"])
	a.Infof("With languages <code>%s</code>", res["MUILanguages"])
	a.Infof("Windows directory is at <code>%s</code>", res["WindowsDirectory"])

	return nil
}

func (a *App) CollectSecurityInfo() error {
	res, err := a.RunWmic(WmicOptions{
		Namespace: "\\\\root\\SecurityCenter2",
		Path:      "AntivirusProduct",
		Columns: []string{
			"displayName",
			"productState",
		},
	})
	if err != nil {
		return errors.WithStack(err)
	}

	if res["displayName"] == "" {
		a.Debugf("No Antivirus product registered with security center.")
		return nil
	}

	a.Infof("Antivirus product: <code>%s</code>", res["displayName"])
	productState, err := strconv.ParseInt(res["productState"], 10, 64)
	if err != nil {
		return errors.WithStack(err)
	}

	var (
		secProviderMask int64 = 0xff0000
		activeMask      int64 = 0x00ff00
		upToDateMask    int64 = 0x0000ff

		secProvider = (productState & secProviderMask) >> 16
		active      = (productState & activeMask) >> 8
		upToDate    = (productState & upToDateMask)
	)

	type ProviderSpec struct {
		Mask  int64
		Label string
	}
	providers := []ProviderSpec{
		{1, "Firewall"},
		{2, "Auto-update Settings"},
		{4, "Antivirus"},
		{8, "Antispyware"},
		{16, "Internet Settings"},
		{32, "User Account Control"},
		{64, "Service"},
	}
	var providerLabels []string
	for _, p := range providers {
		if secProvider&p.Mask > 0 {
			providerLabels = append(providerLabels, fmt.Sprintf("<code>%s</code>", p.Label))
		}
	}

	a.Infof("Provides: %s", strings.Join(providerLabels, ", "))

	if active&16 > 0 {
		a.Infof("Currently active")
	} else {
		a.Infof("Unknown state (can't tell if running)")
	}

	if upToDate&0x10 > 0 {
		a.Warnf("Definitions outdated")
	} else {
		a.Infof("Definitions up-to-date")
	}

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
