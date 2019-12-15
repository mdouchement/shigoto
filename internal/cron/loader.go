package cron

import (
	"path/filepath"

	"github.com/mdouchement/shigoto/pkg/logger"
	"github.com/mdouchement/shigoto/pkg/shigoto"
	"github.com/pkg/errors"
)

// Load loads all shigoto files from the given workdir and registers it to the given pool and starts them.
// It reload already registred shigoto if a change is detected.
func Load(workdir string, pool *Pool, log logger.Logger) error {
	filenames, err := filepath.Glob(filepath.Join(workdir, "*.yml"))
	if err != nil {
		return err
	}

	for _, filename := range filenames {
		shigoto, err := shigoto.Load(filename)
		if err != nil {
			return errors.Wrap(err, filepath.Base(filename))
		}

		if len(shigoto.Baito) == 0 {
			continue
		}

		registred := pool.Get(shigoto.Name)
		if registred == nil {
			pool.Register(shigoto) // New Shigoto
			pool.StartShigoto(shigoto.Name)
			continue
		}

		if shigoto.Same(registred) {
			continue
		}

		// Reloading

		pool.StopShigoto(shigoto.Name)
		pool.Register(shigoto)
		pool.StartShigoto(shigoto.Name)
	}

	return nil
}
