package runner

import (
	"sync"

	"github.com/mdouchement/logger"
)

type base struct {
	sync.Mutex
	ctx         Context
	log         logger.Logger
	ignoreError bool
	deferrable  bool
	err         error
}

func (r *base) IsErrorIgnored() bool {
	return r.ignoreError
}

func (r *base) IsDeferrable() bool {
	return r.deferrable
}

func (r *base) Error() error {
	r.Lock()
	defer r.Unlock()

	return r.err
}

func (r *base) AttachLogger(l logger.Logger) {
	r.Lock()
	defer r.Unlock()

	r.log = l
}
