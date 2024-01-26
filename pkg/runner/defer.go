package runner

import (
	"fmt"

	"github.com/mdouchement/shigoto/pkg/templater"
	"github.com/pkg/errors"
)

type deferrable struct {
	base

	runner Runner
}

func (r *deferrable) Run() {
	r.Lock()
	defer r.Unlock()

	// Reset previous state
	r.err = nil

	//

	r.runner.AttachLogger(r.log.WithPrefix("[defer]"))
	r.runner.Run()
	r.err = r.runner.Error()
}

func init() {
	Register("defer", func(ctx Context, payload map[string]any) (Runner, error) {
		command, ok := payload["defer"]
		if !ok {
			return nil, errors.New("taskfile: defer: missing command value")
		}

		deferrable := &deferrable{
			base: base{
				ctx:        ctx,
				deferrable: true,
			},
		}

		var err error
		templater := templater.New(ctx)

		switch v := command.(type) {
		case string:
			v = templater.Replace(v)
			deferrable.runner, err = lookup(ctx, map[string]any{"exec": v})
		case map[string]any:
			v = templater.ReplaceMapI(v)
			deferrable.runner, err = lookup(ctx, v)
		default:
			return nil, errors.New("invalid command format")
		}

		if deferrable.runner == nil {
			return nil, fmt.Errorf("defer: unknown runner for %v", payload)
		}

		return deferrable, errors.Wrap(err, "defer")
	})
}
