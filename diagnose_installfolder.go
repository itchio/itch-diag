package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/itchio/headway/united"
	"github.com/itchio/lake/tlc"
	"github.com/pkg/errors"
)

type InstallState struct {
	Current string `json:"current"`
	Ready   string `json:"ready"`
}

func (a *App) DiagnoseInstallFolder(installFolder string) error {
	a.Infof("Install folder is <code>%s</code>", installFolder)

	stats, err := os.Stat(installFolder)
	if err != nil {
		a.Errorf("While stat-ing install folder: %+v", err)
		return nil
	}

	if !stats.IsDir() {
		a.Errorf("Install folder is not a directory.")
		return nil
	}

	stateJsonPath := filepath.Join(installFolder, "state.json")

	var installState InstallState
	stateJsonContents, err := ioutil.ReadFile(stateJsonPath)
	if err != nil {
		a.Errorf("While reading install state: %+v", err)
		return nil
	}

	err = json.Unmarshal(stateJsonContents, &installState)
	if err != nil {
		a.Errorf("While decoding install state: %+v", err)
		return nil
	}

	if installState.Current == "" {
		a.Errorf("No current version!")
		return nil
	}
	a.Infof("Install state says current version is <code>%s</code>", installState.Current)

	if installState.Ready != "" {
		a.Warnf("Version <code>%s</code> is ready for update...", installState.Ready)
		return nil
	}

	currentVersionFolder := filepath.Join(installFolder, "app-"+installState.Current)

	container, err := tlc.WalkDir(currentVersionFolder, &tlc.WalkOpts{
		Filter: func(fi os.FileInfo) bool { return true },
	})
	if err != nil {
		return errors.WithStack(err)
	}

	a.Infof("Current version takes up <code>%s</code> in %s", united.FormatBytes(container.Size), container.Stats())
	if container.Size == 0 {
		a.Errorf("Install folder seems empty")
	}

	return nil
}
