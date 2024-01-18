package runner

import (
	"fmt"
	"os"
	osexec "os/exec"
	"time"

	"github.com/gobs/args"
	"github.com/pkg/errors"
)

type exec struct {
	base

	cmd      string
	exec     *osexec.Cmd
	redirect *os.File
}

func (r *exec) Run() {
	r.Lock()
	defer r.Unlock()

	// Reset previous state
	r.err = nil

	//

	start := time.Now()
	logger := r.log.WithPrefixf("[%s]", r.ctx.Name()).WithField("id", GenerateID())
	logger.Info(r.cmd)
	if r.redirect != nil {
		defer r.redirect.Sync()
	}
	if r.ctx.LogsFile() != nil {
		defer r.ctx.LogsFile().Sync()
	}

	err := r.buildCommand()
	if err != nil {
		r.err = err
		logger.WithField("elapsed_time", time.Since(start)).WithField("ignored", r.ignoreError).Error(err)
		return
	}

	if err := r.exec.Run(); err != nil {
		r.err = err
		logger.WithField("elapsed_time", time.Since(start)).WithField("ignored", r.ignoreError).Error(err)
		return
	}

	logger.WithField("elapsed_time", time.Since(start)).Info("finished")
}

func (r *exec) buildCommand() error {
	args := args.GetArgs(r.cmd)
	bin, err := osexec.LookPath(args[0])
	if err != nil {
		return err
	}

	r.exec = &osexec.Cmd{
		Path:   bin,
		Args:   args,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Dir:    r.ctx.Workdir(),
	}

	for k, v := range r.ctx.Environment() {
		r.exec.Env = append(r.exec.Env, fmt.Sprintf("%s=%s", k, v))
	}

	if r.ctx.LogsFile() != nil {
		r.exec.Stdout = r.ctx.LogsFile()
		r.exec.Stderr = r.ctx.LogsFile()
	}

	if r.redirect != nil {
		r.exec.Stdout = r.redirect
		r.exec.Stderr = r.redirect
	}

	return nil
}

func init() {
	Register("exec", func(ctx Context, payload map[string]any) (Runner, error) {
		_, ok := payload["exec"]
		if !ok {
			return nil, errors.New("taskfile: exec: missing command value")
		}

		var err error
		executor := &exec{
			base: base{
				ctx: ctx,
			},
		}

		executor.cmd, ok = payload["exec"].(string)
		if !ok {
			return nil, errors.New("taskfile: exec: command must be a string")
		}
		executor.cmd = executor.ctx.ExpandAll(executor.cmd)
		executor.cmd, err = executor.ctx.ExpandTilde(executor.cmd)
		if err != nil {
			return nil, errors.Wrap(err, "taskfile: exec: could not expand command")
		}

		// Ignore error
		if v, ok := payload["ignore_error"]; ok {
			b, ok := v.(bool)
			if !ok {
				return nil, errors.New("taskfile: exec: ignore_error field must be a boolean")
			}

			executor.ignoreError = b
		}

		// Redirect command stdout/stderr to a file
		if v, ok := payload["redirect"]; ok {
			path, ok := v.(string)
			if !ok {
				return nil, errors.New("taskfile: exec: redirect field must be a string")
			}
			path = executor.ctx.ExpandAll(path)
			path, err = executor.ctx.ExpandTilde(path)
			if err != nil {
				return nil, errors.Wrap(err, "taskfile: exec: could not expand logs redirection file path")
			}

			f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				return nil, errors.Wrap(err, "taskfile: exec: could not create logs redirection file")
			}

			executor.redirect = f
		}

		return executor, nil
	})
}
