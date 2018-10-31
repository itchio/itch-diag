package main

import "github.com/pkg/errors"

// Diagnose runs a battery of tests.
func (a *App) Diagnose() {
	a.Logf("Running diagnostics...")

	a.Eval(`
		window.external.invoke(JSON.stringify({
			UserAgent: window.navigator.userAgent
		}));
	`)
	msg := &UserAgentMessage{}
	a.Receive(&msg)
	a.Logf("User-Agent is: %s", msg.UserAgent)

	a.Must(errors.Errorf("Test crash please ignore. Hi twitter & mastodon friends!"))
	a.Exit()
}
