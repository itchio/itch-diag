package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/zserge/webview"
)

// App contains all the state for itch diag
type App struct {
	w     webview.WebView
	queue chan string
}

const ItchDiagVersion = "0.3.0"

func main() {
	queue := make(chan string, 20)
	w := webview.New(webview.Settings{
		URL:       `data:text/html,` + url.PathEscape(baseHTML),
		Title:     fmt.Sprintf("itch diagnostics v%s", ItchDiagVersion),
		Width:     1100,
		Height:    800,
		Resizable: true,
		ExternalInvokeCallback: func(w webview.WebView, payload string) {
			queue <- payload
		},
	})
	w.InjectCSS(baseCSS)

	app := &App{
		w:     w,
		queue: queue,
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				app.Errorf("Recovered from panic: %#v", r)
				app.Errorf("<pre>%+v</pre>", errors.Errorf("stack trace"))
			}
		}()
		app.Diagnose()
	}()

	app.Run()
}

type LogGroup interface {
	Item(format string, args ...interface{}) LogGroup
	End()
}

type logGroup struct {
	level string
	a     *App
	items []string
}

func (a *App) InfoGroup() LogGroup {
	return a.Group("info")
}

func (a *App) Group(level string) LogGroup {
	return &logGroup{a: a, level: level, items: nil}
}

func (lg *logGroup) Item(format string, args ...interface{}) LogGroup {
	lg.items = append(lg.items, fmt.Sprintf(format, args...))
	return lg
}

func (lg *logGroup) End() {
	lg.a.Logf(lg.level, "%s", strings.Join(lg.items, " — "))
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

	log.Print(line)
	a.w.Dispatch(func() {
		err := a.w.Eval(`
			(function () {
				var p = document.createElement("p");
				p.className = "level-` + level + `";
				p.innerHTML = ` + string(payload) + `;
				document.querySelector("#app").appendChild(p);
			})()
		`)
		if err != nil {
			panic(err)
		}
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
		err := a.w.Eval(`(function() {` + code + `})()`)
		if err != nil {
			panic(err)
		}
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
