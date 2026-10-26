// Package runtime runs control scripts with the embedded mvdan.cc/sh
// interpreter. It is a minimal consumer for now; the script-runtime change
// reshapes it.
package runtime

import (
	"bytes"
	"context"
	"errors"
	"strings"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

// Result is the outcome of a finished script.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Run runs script in dir and returns its output and exit code. A non-zero
// exit code is not an error; an error means the script could not be parsed
// or the interpreter failed.
func Run(ctx context.Context, script, dir string) (Result, error) {
	file, err := syntax.NewParser().Parse(strings.NewReader(script), "script.sh")
	if err != nil {
		return Result{}, err
	}

	var stdout, stderr bytes.Buffer
	runner, err := interp.New(
		interp.Dir(dir),
		interp.StdIO(nil, &stdout, &stderr),
		interp.ExecHandlers(utilitiesHandler),
	)
	if err != nil {
		return Result{}, err
	}

	res := Result{}
	err = runner.Run(ctx, file)
	var status interp.ExitStatus
	switch {
	case errors.As(err, &status):
		res.ExitCode = int(status)
	case err != nil:
		return Result{}, err
	}
	res.Stdout = stdout.String()
	res.Stderr = stderr.String()
	return res, nil
}
