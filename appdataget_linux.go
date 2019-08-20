//+build linux

package main

import (
	"os"
	"path/filepath"
)

func (a *App) GetAppDataFolder() (string, error) {
	appName := "itch"

	configPath := os.Getenv("XDG_CONFIG_HOME")
	if configPath != "" {
		return filepath.Join(configPath, appName), nil
	} else {
		homePath := os.Getenv("HOME")
		return filepath.Join(homePath, ".config", appName), nil
	}	
}

