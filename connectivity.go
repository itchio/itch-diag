package main

import (
	"net/http"
	"time"

	"github.com/itchio/httpkit/timeout"
)

func (a *App) DiagnoseConnectivity() error {
	a.TestEndpoint("https://static.itch.io/ping.txt")
	a.TestEndpoint("https://itch.io/static/ping.txt")
	a.TestEndpoint("https://broth.itch.ovh")
	a.TestEndpoint("http://locales.itch.ovh/itch/en.json")

	return nil
}

func (a *App) TestEndpoint(endpoint string) {
	startTime := time.Now()

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		a.Errorf("<code>%s</code>: %+v", endpoint)
		return
	}

	client := timeout.NewDefaultClient()
	res, err := client.Do(req)
	if err != nil {
		a.Errorf("<code>%s</code>: %+v", endpoint)
		return
	}
	defer res.Body.Close()

	a.Infof("<code>%s</code>: HTTP %d (in %s)", endpoint, res.StatusCode, time.Since(startTime))
}
