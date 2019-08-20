package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/itchio/headway/united"
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
	a.Infof("broth packages: %s", brothPackages)

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

	butlerExecutable := filepath.Join(butlerChosenFolder, "butler")
	if runtime.GOOS == "windows" {
		butlerExecutable += ".exe"
	}
	a.Debugf("Verifying <code>%s</code>", butlerExecutable)

	var timeout = 5 * time.Second

	{
		a.Debugf("Retrieving butler version...")
		errs := make(chan error)
		ctx, cancel := context.WithCancel(context.Background())
		var butlerVersion string

		retrieveButlerVersion := func() error {
			out, err := exec.Command(butlerExecutable, "-V").CombinedOutput()
			if err != nil {
				return errors.WithStack(err)
			}
			butlerVersion = strings.TrimSpace(string(out))
			return nil
		}

		go func() {
			timer := time.After(timeout)
			select {
			case <-ctx.Done():
			case <-timer:
				errs <- errors.Errorf("Timed out after %s", timeout)
			}
		}()
		go func() {
			defer cancel()
			errs <- retrieveButlerVersion()
		}()

		err := <-errs
		if err != nil {
			return errors.WithStack(err)
		}
		a.Infof("butler version: <code>%s</code>", butlerVersion)
	}

	err = a.TestButlerd(appDataFolder, butlerExecutable)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (a *App) ListFiles(folder string) (string, error) {
	items, err := ioutil.ReadDir(folder)
	if err != nil {
		return "", errors.WithStack(err)
	}

	var names []string
	for _, item := range items {
		suffix := fmt.Sprintf(" (%s)", united.FormatBytes(item.Size()))
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

func (a *App) EnsureFile(file string) error {
	stats, err := os.Stat(file)
	if err != nil {
		return errors.WithStack(err)
	}

	if stats.IsDir() {
		return errors.Errorf("%s: should be a file", file)
	}

	return nil
}
