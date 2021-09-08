package runner

import (
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"github.com/traefik/yaegi/stdlib/unrestricted"
)

type yaegi struct {
	base

	src string
}

func (r *yaegi) Run() {
	r.Lock()
	defer r.Unlock()

	// Reset previous state
	r.err = nil

	//

	start := time.Now()
	logger := r.log.WithField("prefix", r.ctx.Name()).WithField("id", GenerateID())
	logger.Info("Running Yaegi script")
	if r.ctx.LogsFile() != nil {
		defer r.ctx.LogsFile().Sync()
	}

	i := interp.New(interp.Options{})
	// i.Use(syscall.Symbols)
	// i.Use(unsafe.Symbols)
	i.Use(stdlib.Symbols)
	i.Use(unrestricted.Symbols)
	i.Use(interp.Exports{
		"logger/logger": { // Use as `import "logger"`
			"WithPrefix":  reflect.ValueOf(r.log.WithPrefix),
			"WithPrefixf": reflect.ValueOf(r.log.WithPrefixf),
			"WithField":   reflect.ValueOf(r.log.WithField),
			"WithError":   reflect.ValueOf(r.log.WithError),
			"WithFields":  reflect.ValueOf(r.log.WithFields),
			//
			"Debug":  reflect.ValueOf(r.log.Debug),
			"Debugf": reflect.ValueOf(r.log.Debugf),
			"Info":   reflect.ValueOf(r.log.Info),
			"Infof":  reflect.ValueOf(r.log.Infof),
			"Warn":   reflect.ValueOf(r.log.Warn),
			"Warnf":  reflect.ValueOf(r.log.Warnf),
			"Error":  reflect.ValueOf(r.log.Error),
			"Errorf": reflect.ValueOf(r.log.Errorf),
			//
			"Print":   reflect.ValueOf(r.log.Print),
			"Printf":  reflect.ValueOf(r.log.Printf),
			"Println": reflect.ValueOf(r.log.Println),
			"Fatal":   reflect.ValueOf(r.log.Fatal),
			"Fatalf":  reflect.ValueOf(r.log.Fatalf),
			"Fatalln": reflect.ValueOf(r.log.Fatalln),
			"Panic":   reflect.ValueOf(r.log.Panic),
			"Panicf":  reflect.ValueOf(r.log.Panicf),
			"Panicln": reflect.ValueOf(r.log.Panicln),
		},
	})

	if _, err := i.Eval(r.src); err != nil {
		r.err = err
		logger.WithField("elapsed_time", time.Since(start)).WithField("ignored", r.ignoreError).Error(err)
		return
	}

	logger.WithField("elapsed_time", time.Since(start)).Info("finished")
}

func init() {
	Register("yaegi", func(ctx Context, payload map[string]interface{}) (Runner, error) {
		_, ok := payload["yaegi"]
		if !ok {
			return nil, errors.New("taskfile: yaegi: missing src value")
		}

		executor := &yaegi{
			base: base{
				ctx: ctx,
			},
		}

		executor.src, ok = payload["yaegi"].(string)
		if !ok {
			return nil, errors.New("taskfile: yaegi: src must be a string")
		}

		// Check if src is a file and not plain code.
		if strings.HasSuffix(strings.TrimSpace(executor.src), ".go") {
			src, err := os.ReadFile(executor.ctx.ExpandAll(executor.src))
			if err != nil {
				return nil, errors.Wrap(err, "taskfile: yaegi: file")
			}
			executor.src = string(src)
		}

		executor.src = executor.ctx.ExpandAll(executor.src)

		// Ignore error
		if v, ok := payload["ignore_error"]; ok {
			b, ok := v.(bool)
			if !ok {
				return nil, errors.New("taskfile: yaegi: ignore_error field must be a boolean")
			}

			executor.ignoreError = b
		}

		return executor, nil
	})
}
