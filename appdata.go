package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/itchio/httpkit/progress"
	"github.com/pkg/errors"
)

func (a *App) DiagnoseAppData() error {
	appDataFolder, err := a.GetAppDataFolder()
	if err != nil {
		return errors.WithStack(err)
	}

	err = a.EnsureFolder(appDataFolder)
	if err != nil {
		return errors.WithStack(err)
	}

	brothFolder := filepath.Join(appDataFolder, "broth")
	err = a.EnsureFolder(brothFolder)
	if err != nil {
		return errors.WithStack(err)
	}

	brothPackages, err := a.ListFiles(brothFolder)
	if err != nil {
		return errors.WithStack(err)
	}
	a.Infof("Broth packages: %s", brothPackages)

	butlerFolder := filepath.Join(brothFolder, "butler")
	err = a.EnsureFolder(butlerFolder)
	if err != nil {
		return errors.WithStack(err)
	}

	butlerVersionsFolder := filepath.Join(butlerFolder, "versions")
	butlerVersions, err := a.ListFiles(butlerVersionsFolder)
	if err != nil {
		return errors.WithStack(err)
	}
	a.Infof("butler versions: %s", butlerVersions)

	butlerChosenVersionPath := filepath.Join(butlerFolder, ".chosen-version")
	butlerChosenVersionContents, err := ioutil.ReadFile(butlerChosenVersionPath)
	if err != nil {
		return errors.WithStack(err)
	}
	butlerChosenVersion := string(butlerChosenVersionContents)
	a.Infof("butler chosen version: <code>%s</code>", butlerChosenVersion)

	butlerChosenFolder := filepath.Join(butlerVersionsFolder, butlerChosenVersion)

	butlerInstallMarker := filepath.Join(butlerChosenFolder, ".installed")
	butlerInstallMarkerContents, err := ioutil.ReadFile(butlerInstallMarker)
	if err != nil {
		return errors.WithStack(err)
	}
	a.Infof("Install marker: <code>%s</code>", string(butlerInstallMarkerContents))

	butlerChosenFiles, err := a.ListFiles(butlerChosenFolder)
	if err != nil {
		return errors.WithStack(err)
	}
	a.Infof("Installed files: %s", butlerChosenFiles)

	return nil
}

func (a *App) ListFiles(folder string) (string, error) {
	items, err := ioutil.ReadDir(folder)
	if err != nil {
		return "", errors.WithStack(err)
	}

	var names []string
	for _, item := range items {
		suffix := fmt.Sprintf(" (%s)", progress.FormatBytes(item.Size()))
		if item.IsDir() {
			suffix = "/"
		}

		names = append(names, fmt.Sprintf("<code>%s%s</code>",
			item.Name(),
			suffix,
		))
	}
	return strings.Join(names, ", "), nil
}

func (a *App) EnsureFolder(folder string) error {
	stats, err := os.Stat(folder)
	if err != nil {
		return errors.WithStack(err)
	}

	if !stats.IsDir() {
		return errors.Errorf("%s: should be a folder", folder)
	}

	return nil
}
