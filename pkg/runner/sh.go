package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

type sh struct {
	base

	script   string
	file     *syntax.File
	redirect *os.File
}

func (r *sh) Run() {
	r.Lock()
	defer r.Unlock()

	// Reset previous state
	r.err = nil

	//

	start := time.Now()
	logger := r.log.WithField("prefix", r.ctx.Name()).WithField("id", GenerateID())
	logger.Info(strings.Split(r.script, "\n")[0] + "...")
	if r.redirect != nil {
		defer r.redirect.Sync()
	}
	if r.ctx.LogsFile() != nil {
		defer r.ctx.LogsFile().Sync()
	}

	shell, err := r.buildShell()
	if err != nil {
		r.err = err
		logger.WithField("elapsed_time", time.Since(start)).WithField("ignored", r.ignoreError).Error(err)
		return
	}

	if err := shell.Run(context.Background(), r.file); err != nil {
		r.err = err
		logger.WithField("elapsed_time", time.Since(start)).WithField("ignored", r.ignoreError).Error(err)
		return
	}

	logger.WithField("elapsed_time", time.Since(start)).Info("finished")
}

func (r *sh) buildShell() (*interp.Runner, error) {
	environ := os.Environ()
	for k, v := range r.ctx.Environment() {
		environ = append(environ, fmt.Sprintf("%s=%s", k, v))
	}

	stdio := interp.StdIO(os.Stdin, os.Stdout, os.Stderr)
	if r.ctx.LogsFile() != nil {
		stdio = interp.StdIO(os.Stdin, r.ctx.LogsFile(), r.ctx.LogsFile())
	}

	if r.redirect != nil {
		stdio = interp.StdIO(os.Stdin, r.redirect, r.redirect)
	}

	return interp.New(
		interp.Dir(r.ctx.Workdir()),
		interp.Env(expand.ListEnviron(environ...)),

		interp.OpenHandler(func(ctx context.Context, path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
			if path == "/dev/null" {
				return devNull{}, nil
			}
			return interp.DefaultOpenHandler()(ctx, path, flag, perm)
		}),

		stdio,
	)
}

func init() {
	Register("sh", func(ctx Context, payload map[string]interface{}) (Runner, error) {
		_, ok := payload["sh"]
		if !ok {
			return nil, errors.New("taskfile: sh: missing script value")
		}

		executor := &sh{
			base: base{
				ctx: ctx,
			},
		}

		executor.script, ok = payload["sh"].(string)
		if !ok {
			return nil, errors.New("taskfile: sh: command must be a string")
		}
		executor.script = executor.ctx.ExpandVariables(executor.script)

		var err error
		executor.file, err = syntax.NewParser().Parse(strings.NewReader(executor.script), "")
		if err != nil {
			return nil, errors.Wrap(err, "taskfile: sh: parse script")
		}

		// Ignore error
		if v, ok := payload["ignore_error"]; ok {
			b, ok := v.(bool)
			if !ok {
				return nil, errors.New("taskfile: sh: ignore_error field must be a boolean")
			}

			executor.ignoreError = b
		}

		// Redirect command stdout/stderr to a file
		if v, ok := payload["redirect"]; ok {
			path, ok := v.(string)
			if !ok {
				return nil, errors.New("taskfile: sh: redirect field must be a string")
			}
			path = executor.ctx.ExpandAll(path)

			f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, errors.Wrap(err, "could not create logs redirection file")
			}

			executor.redirect = f
		}

		return executor, nil
	})
}

// ----------------
// -------------
// /dev/null
// -----
// ---

// https://github.com/go-task/task/blob/master/internal/execext/devnull.go

var _ io.ReadWriteCloser = devNull{}

type devNull struct{}

func (devNull) Read(p []byte) (int, error)  { return 0, io.EOF }
func (devNull) Write(p []byte) (int, error) { return len(p), nil }
func (devNull) Close() error                { return nil }
