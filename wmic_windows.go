package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type WmicResult map[string]string

type WmicOptions struct {
	Alias     string
	Namespace string
	Path      string

	Columns []string
}

func (a *App) RunWmic(opts WmicOptions) (WmicResult, error) {
	var args []string
	if opts.Alias != "" {
		args = append(args, opts.Alias)
	} else {
		if opts.Namespace == "" || opts.Path == "" {
			err := errors.Errorf("Internal error: either Alias or Namespace+Path need to be specified for wmic")
			return nil, err
		}
		args = append(args, fmt.Sprintf("/namespace:%s", opts.Namespace))
		args = append(args, "path")
		args = append(args, opts.Path)
	}

	args = append(args, "get")
	for i, column := range opts.Columns {
		last := i == len(opts.Columns)-1
		if last {
			args = append(args, column)
		} else {
			args = append(args, column+",")
		}
	}
	args = append(args, "/format:list")

	out, err := exec.Command("wmic", args...).CombinedOutput()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	text := string(out)
	text = strings.Replace(text, "\r\n", "\n", -1)
	text = strings.TrimSpace(text)

	result := make(WmicResult)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		tokens := strings.SplitN(line, "=", 2)
		if len(tokens) == 2 {
			key, value := tokens[0], tokens[1]
			result[key] = value
		}
	}

	return result, nil
}
