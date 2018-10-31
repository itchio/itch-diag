//+build windows

package main

import (
	"path/filepath"

	"github.com/itchio/ox/winox"
	"github.com/pkg/errors"
)

func (a *App) GetAppDataFolder() (string, error) {
	base, err := winox.GetFolderPath(winox.FolderTypeAppData)
	if err != nil {
		return "", errors.WithStack(err)
	}

	itchFolder := filepath.Join(base, "itch")
	return itchFolder, nil
}
