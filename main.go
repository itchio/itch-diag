package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/zserge/webview"
)

// App contains all the state for itch diag
type App struct {
	w     webview.WebView
	queue chan string
}

func main() {
	queue := make(chan string, 20)
	w := webview.New(webview.Settings{
		URL:    `data:text/html,` + url.PathEscape(baseHTML),
		Title:  "itch diagnostics",
		Width:  1200,
		Height: 600,
		ExternalInvokeCallback: func(w webview.WebView, payload string) {
			queue <- payload
		},
	})
	w.InjectCSS(baseCSS)

	app := &App{
		w:     w,
		queue: queue,
	}

	go app.Diagnose()

	app.Run()
}

func (a *App) Debugf(format string, args ...interface{}) {
	a.Logf("debug", format, args...)
}

func (a *App) Successf(format string, args ...interface{}) {
	a.Logf("success", format, args...)
}

func (a *App) Infof(format string, args ...interface{}) {
	a.Logf("info", format, args...)
}

func (a *App) Warnf(format string, args ...interface{}) {
	a.Logf("warn", format, args...)
}

func (a *App) Errorf(format string, args ...interface{}) {
	a.Logf("error", format, args...)
}

func (a *App) Logf(level string, format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	payload, err := json.Marshal(line)
	if err != nil {
		panic(err)
	}

	a.w.Dispatch(func() {
		a.w.Eval(`
			(function () {
				var p = document.createElement("p");
				p.className = "level-` + level + `";
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

func (a *App) Receive(dst interface{}) {
	payload := <-a.queue
	err := json.Unmarshal([]byte(payload), &dst)
	a.Must(err)
}

func (a *App) Must(err error) {
	if err != nil {
		a.w.Dialog(
			webview.DialogTypeAlert,
			0,
			"Fatal error",
			fmt.Sprintf("fatal error: %+v", err),
		)
		os.Exit(1)
	}
}

func (a *App) Eval(code string) {
	a.w.Dispatch(func() {
		a.w.Eval(`(function() {` + code + `})()`)
	})
}

func (a *App) Test(label string, run func() error) {
	a.Debugf("%s...", label)

	err := run()
	if err != nil {
		a.Warnf("While doing '%s': <pre>%+v</pre>", label, err)
		return
	}
}

func (a *App) Exit() {
	a.w.Exit()
}
