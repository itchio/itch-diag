package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/zserge/webview"
)

// App contains all the state for itch diag
type App struct {
	w webview.WebView
}

func main() {
	w := webview.New(webview.Settings{
		URL:    `data:text/html,` + url.PathEscape(baseHTML),
		Title:  "itch diagnostics",
		Width:  800,
		Height: 600,
	})
	app := &App{
		w: w,
	}

	go func() {
		for {
			time.Sleep(1 * time.Second)
			app.Logf("Current time: <i>%s</i>", time.Now().Format(time.RFC3339))
		}
	}()

	app.Run()
}

// Logf prints a line in the app view
func (a *App) Logf(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	payload, err := json.Marshal(line)
	if err != nil {
		panic(err)
	}

	a.w.Dispatch(func() {
		a.w.Eval(`
			(function () {
				var p = document.createElement("p");
				p.innerHTML = ` + string(payload) + `;
				document.querySelector("#app").appendChild(p);
			})()
		`)
	})
}

// Run blocks until the webview closes
func (a *App) Run() {
	a.w.Run()
}
