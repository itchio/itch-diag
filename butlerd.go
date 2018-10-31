package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/itchio/httpkit/timeout"
	"github.com/pkg/errors"
)

func (a *App) TestButlerd(appDataFolder string, butlerExecutable string) error {
	dbPath := filepath.Join(appDataFolder, "db", "butler.db")
	err := a.EnsureFile(dbPath)
	if err != nil {
		return errors.WithStack(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var addr string
	var secret string
	addrErrs := make(chan error)
	addrCtx, addrCancel := context.WithCancel(ctx)

	cmd := exec.CommandContext(
		ctx,
		"butler",
		"--json",
		"--dbpath", dbPath,
		"daemon",
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.WithStack(err)
	}
	go a.relay(stdout, "butler stdout", func(line string) bool {
		msg := make(map[string]interface{})

		err := json.Unmarshal([]byte(line), &msg)
		if err != nil {
			// skip non-JSON stuff
			return false
		}

		typ, ok := msg["type"]
		if !ok {
			return false
		}

		if typ.(string) == "butlerd/listen-notification" {
			http := msg["http"].(map[string]interface{})
			secret = msg["secret"].(string)
			addr = http["address"].(string)
			addrErrs <- nil
			addrCancel()
			return true
		}
		return false
	})

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return errors.WithStack(err)
	}
	go a.relay(stderr, "butler stderr", nil)

	a.Debugf("Starting butler daemon...")
	err = cmd.Start()
	if err != nil {
		return errors.WithStack(err)
	}

	a.Debugf("Waiting for daemon address")
	go func() {
		timeout := 5 * time.Second
		timer := time.After(timeout)
		select {
		case <-timer:
			addrErrs <- errors.Errorf("Timed out after %s", timeout)
		case <-addrCtx.Done():
			// all good!
		}
	}()

	err = <-addrErrs
	if err != nil {
		return errors.WithStack(err)
	}

	a.Debugf("Daemon is listening on <code>%s</code>", addr)

	type RPCReply struct {
		ID     int64            `json:"id"`
		Result *json.RawMessage `json:"result"`
	}

	type User struct {
		DisplayName string `json:"displayName"`
	}

	type Profile struct {
		ID            int64  `json:"id"`
		User          User   `json:"user"`
		LastConnected string `json:"lastConnected"`
	}

	type ListProfilesResponse struct {
		Profiles []Profile `json:"profiles"`
	}

	var reqID int64 = 0
	client := timeout.NewDefaultClient()

	testCall := func() error {
		a.Debugf("Listing profiles...")
		reqAddr := "http://" + addr + "/call/Profile.List"
		reqBody := strings.NewReader("{}")
		req, err := http.NewRequest("POST", reqAddr, reqBody)
		if err != nil {
			return errors.WithStack(err)
		}

		reqID++
		req.Header.Set("X-ID", fmt.Sprintf("%d", reqID))
		req.Header.Set("X-Secret", secret)

		res, err := client.Do(req)
		if err != nil {
			return errors.WithStack(err)
		}
		defer res.Body.Close()

		payload, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return errors.WithStack(err)
		}

		var reply RPCReply
		err = json.Unmarshal(payload, &reply)
		if err != nil {
			return errors.WithStack(err)
		}

		var profs ListProfilesResponse
		err = json.Unmarshal(*reply.Result, &profs)
		if err != nil {
			return errors.WithStack(err)
		}

		a.Infof("Found %d profiles", len(profs.Profiles))
		for _, p := range profs.Profiles {
			a.Infof("- %s", p.User.DisplayName)
		}

		return nil
	}
	err = testCall()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (a *App) formatJSON(input []byte) string {
	inter := make(map[string]interface{})

	err := json.Unmarshal(input, &inter)
	a.Must(err)

	res, err := json.MarshalIndent(inter, "", "  ")
	a.Must(err)
	return string(res)
}

func (a *App) relay(reader io.Reader, label string, processLine func(string) bool) {
	s := bufio.NewScanner(reader)
	for s.Scan() {
		line := s.Text()
		if processLine != nil {
			if processLine(line) {
				continue
			}
		}
		a.Debugf("[%s] %s", label, line)
	}
}
