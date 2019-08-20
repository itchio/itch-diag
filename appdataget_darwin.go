//+build darwin

package main

import "github.com/pkg/errors"

func (a *App) GetAppDataFolder() (string, error) {
	return "", errors.Errorf("stub!")
}
