package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	"os/exec"
	"path/filepath"
	"time"

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
	defer addrCancel()

	cmd := exec.CommandContext(
		ctx,
		butlerExecutable,
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
			tcp := msg["tcp"].(map[string]interface{})
			secret = msg["secret"].(string)
			addr = tcp["address"].(string)
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

	type RPCRequest struct {
		JSONRPC string      `json:"jsonrpc"`
		ID      int64       `json:"id"`
		Method  string      `json:"method"`
		Params  interface{} `json:"params"`
	}

	type RPCReply struct {
		ID     int64                  `json:"id"`
		Result *json.RawMessage       `json:"result"`
		Error  map[string]interface{} `json:"error"`
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
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return errors.WithStack(err)
	}
	a.Debugf("Connected...")

	sendReq := func(req RPCRequest) error {
		reqBytes, err := json.Marshal(req)
		if err != nil {
			return errors.WithStack(err)
		}
		reqID++

		_, err = conn.Write(reqBytes)
		if err != nil {
			return errors.WithStack(err)
		}
		_, err = conn.Write([]byte{'\n'})
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	readReply := func() (*RPCReply, error) {
		s := bufio.NewScanner(conn)
		if !s.Scan() {
			return nil, errors.Errorf("expected to read a line from TCP socket, but didn't")
		}

		var reply RPCReply
		err = json.Unmarshal(s.Bytes(), &reply)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if reply.Error != nil {
			return nil, errors.Errorf("JSON-RPC error: %#v", reply.Error)
		}

		return &reply, nil
	}

	testCall := func() error {
		{
			a.Debugf("Authenticating...")
			req := RPCRequest{
				JSONRPC: "2.0",
				ID:      reqID,
				Method:  "Meta.Authenticate",
				Params: map[string]interface{}{
					"secret": secret,
				},
			}
			reqID++

			err = sendReq(req)
			if err != nil {
				return errors.WithStack(err)
			}

			_, err := readReply()
			if err != nil {
				return errors.WithStack(err)
			}
		}

		a.Debugf("Listing profiles...")
		req := RPCRequest{
			JSONRPC: "2.0",
			ID:      reqID,
			Method:  "Profile.List",
			Params:  nil,
		}
		err = sendReq(req)
		if err != nil {
			return errors.WithStack(err)
		}

		reply, err := readReply()
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
