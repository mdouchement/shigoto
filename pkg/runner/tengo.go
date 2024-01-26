package runner

import (
	"os"
	"strings"
	"time"

	tengopkg "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/mdouchement/ldt/pkg/tengolib"
	"github.com/mdouchement/logger"
	"github.com/pkg/errors"
)

type tengo struct {
	base

	src      string
	redirect *os.File
}

func (r *tengo) Run() {
	r.Lock()
	defer r.Unlock()

	// Reset previous state
	r.err = nil

	//

	start := time.Now()
	logger := r.log.WithPrefixf("[%s]", r.ctx.Name()).WithField("id", GenerateID())
	logger.Info("Running tengo script")

	// Load modules
	modules := stdlib.GetModuleMap(stdlib.AllModuleNames()...)
	tengolib.MergeModule(modules, tengolib.AllModuleNames()...)
	modules.AddBuiltinModule("shigoto", map[string]tengopkg.Object{
		"logger": withlogger(logger),
	})

	// Compile source code
	script := tengopkg.NewScript([]byte(r.src))
	script.SetImports(modules)
	if _, err := script.Run(); err != nil {
		r.err = err
		logger.WithField("elapsed_time", time.Since(start)).WithField("ignored", r.ignoreError).Error(err)
		return
	}

	logger.WithField("elapsed_time", time.Since(start)).Info("finished")
}

func init() {
	Register("tengo", func(ctx Context, payload map[string]interface{}) (Runner, error) {
		_, ok := payload["tengo"]
		if !ok {
			return nil, errors.New("taskfile: tengo: missing src value")
		}

		executor := &tengo{
			base: base{
				ctx: ctx,
			},
		}

		executor.src, ok = payload["tengo"].(string)
		if !ok {
			return nil, errors.New("taskfile: tengo: src must be a string")
		}

		// Check if src is a file and not plain code.
		if strings.HasSuffix(strings.TrimSpace(executor.src), ".tengo") || strings.HasSuffix(strings.TrimSpace(executor.src), ".tgo") {
			executor.src = executor.ctx.ExpandAll(executor.src)
			var err error
			executor.src, err = executor.ctx.ExpandTilde(executor.src)
			if !ok {
				return nil, errors.Wrap(err, "taskfile: tengo: expand filename")
			}

			src, err := os.ReadFile(executor.src)
			if err != nil {
				return nil, errors.Wrap(err, "taskfile: tengo: file")
			}
			executor.src = string(src)
		}

		executor.src = executor.ctx.ExpandAll(executor.src)
		executor.src = "shigoto := import(\"shigoto\")\n" + executor.src // Pre-import shigoto.

		// Ignore error
		if v, ok := payload["ignore_error"]; ok {
			b, ok := v.(bool)
			if !ok {
				return nil, errors.New("taskfile: tengo: ignore_error field must be a boolean")
			}

			executor.ignoreError = b
		}

		return executor, nil
	})
}

func withlogger(logger logger.Logger) *tengopkg.ImmutableMap {
	return &tengopkg.ImmutableMap{
		Value: map[string]tengopkg.Object{
			"with_prefix": &tengopkg.UserFunction{
				Name: "with_prefix",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) != 1 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					s1, ok := tengopkg.ToString(args[0])
					if !ok {
						return nil, tengopkg.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "string(compatible)",
							Found:    args[0].TypeName(),
						}
					}

					l := logger.WithPrefix(s1)
					return withlogger(l), nil
				},
			},
			"with_prefixf": &tengopkg.UserFunction{
				Name: "with_prefixf",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) < 2 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					str, err := tengolib.Format(args...)
					if err != nil {
						return nil, err
					}

					l := logger.WithPrefix(str)
					return withlogger(l), nil
				},
			},
			"with_error": &tengopkg.UserFunction{
				Name: "with_error",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) != 1 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					terr, ok := args[0].(*tengopkg.Error)
					if !ok {
						return nil, tengopkg.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "error(compatible)",
							Found:    args[0].TypeName(),
						}
					}

					m, _ := tengopkg.ToString(terr.Value)
					l := logger.WithError(errors.New(m))
					return withlogger(l), nil
				},
			},
			"with_field": &tengopkg.UserFunction{
				Name: "with_field",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) != 2 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					k, ok := tengopkg.ToString(args[0])
					if !ok {
						return nil, tengopkg.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "string(compatible)",
							Found:    args[0].TypeName(),
						}
					}

					l := logger.WithField(k, tengopkg.ToInterface(args[1]))
					return withlogger(l), nil
				},
			},
			"with_fields": &tengopkg.UserFunction{
				Name: "with_fields",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) != 1 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					tm, ok := args[0].(*tengopkg.Map)
					if !ok {
						return nil, tengopkg.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "map(compatible)",
							Found:    args[0].TypeName(),
						}
					}

					m := make(map[string]any)
					for k, v := range tm.Value {
						m[k] = tengopkg.ToInterface(v)
					}

					l := logger.WithFields(m)
					return withlogger(l), nil
				},
			},
			//
			//
			//
			"debug": &tengopkg.UserFunction{
				Name: "debug",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) < 1 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					logger.Debug(tengolib.InterfaceArray(args)...)
					return tengopkg.UndefinedValue, nil
				},
			},
			"debugf": &tengopkg.UserFunction{
				Name: "debugf",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) < 2 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					str, err := tengolib.Format(args...)
					if err != nil {
						return nil, err
					}

					logger.Debug(str)
					return tengopkg.UndefinedValue, nil
				},
			},
			"info": &tengopkg.UserFunction{
				Name: "info",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) < 1 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					logger.Info(tengolib.InterfaceArray(args)...)
					return tengopkg.UndefinedValue, nil
				},
			},
			"infof": &tengopkg.UserFunction{
				Name: "infof",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) < 2 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					str, err := tengolib.Format(args...)
					if err != nil {
						return nil, err
					}

					logger.Info(str)
					return tengopkg.UndefinedValue, nil
				},
			},
			"warn": &tengopkg.UserFunction{
				Name: "warn",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) < 1 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					logger.Warn(tengolib.InterfaceArray(args)...)
					return tengopkg.UndefinedValue, nil
				},
			},
			"warnf": &tengopkg.UserFunction{
				Name: "warnf",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) < 2 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					str, err := tengolib.Format(args...)
					if err != nil {
						return nil, err
					}

					logger.Warnf(str)
					return tengopkg.UndefinedValue, nil
				},
			},
			"error": &tengopkg.UserFunction{
				Name: "error",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) < 1 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					logger.Error(tengolib.InterfaceArray(args)...)
					return tengopkg.UndefinedValue, nil
				},
			},
			"errorf": &tengopkg.UserFunction{
				Name: "errorf",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) < 2 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					str, err := tengolib.Format(args...)
					if err != nil {
						return nil, err
					}

					logger.Errorf(str)
					return tengopkg.UndefinedValue, nil
				},
			},
			//
			//
			//
			"print": &tengopkg.UserFunction{
				Name: "print",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) < 1 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					logger.Print(tengolib.InterfaceArray(args)...)
					return tengopkg.UndefinedValue, nil
				},
			},
			"printf": &tengopkg.UserFunction{
				Name: "printf",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) < 2 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					str, err := tengolib.Format(args...)
					if err != nil {
						return nil, err
					}

					logger.Print(str)
					return tengopkg.UndefinedValue, nil
				},
			},
			"println": &tengopkg.UserFunction{
				Name: "println",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) < 1 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					logger.Println(tengolib.InterfaceArray(args)...)
					return tengopkg.UndefinedValue, nil
				},
			},
			"fatal": &tengopkg.UserFunction{
				Name: "fatal",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) < 1 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					logger.Fatal(tengolib.InterfaceArray(args)...)
					return tengopkg.UndefinedValue, nil
				},
			},
			"fatalf": &tengopkg.UserFunction{
				Name: "fatalf",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) < 2 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					str, err := tengolib.Format(args...)
					if err != nil {
						return nil, err
					}

					logger.Fatal(str)
					return tengopkg.UndefinedValue, nil
				},
			},
			"fatalln": &tengopkg.UserFunction{
				Name: "fatalln",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) < 1 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					logger.Println(tengolib.InterfaceArray(args)...)
					return tengopkg.UndefinedValue, nil
				},
			},
			"panic": &tengopkg.UserFunction{
				Name: "panic",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) < 1 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					logger.Panic(tengolib.InterfaceArray(args)...)
					return tengopkg.UndefinedValue, nil
				},
			},
			"panicf": &tengopkg.UserFunction{
				Name: "panicf",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) < 2 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					str, err := tengolib.Format(args...)
					if err != nil {
						return nil, err
					}

					logger.Panic(str)
					return tengopkg.UndefinedValue, nil
				},
			},
			"panicln": &tengopkg.UserFunction{
				Name: "panicln",
				Value: func(args ...tengopkg.Object) (tengopkg.Object, error) {
					if len(args) < 1 {
						return nil, tengopkg.ErrWrongNumArguments
					}

					logger.Panicln(tengolib.InterfaceArray(args)...)
					return tengopkg.UndefinedValue, nil
				},
			},
		},
	}
}
