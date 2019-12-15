package runner

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/mdouchement/shigoto/pkg/io"
	"github.com/mdouchement/shigoto/pkg/logger"
)

type (
	// A Runner is the action to be executed.
	Runner interface {
		Run()
		Error() error
		IsErrorIgnored() bool
		AttachLogger(l logger.Logger)
	}

	// A Context carries the context of a Runner.
	Context interface {
		Name() string
		Environment() map[string]string
		ExpandAll(string) string
		ExpandVariables(string) string
		Workdir() string
		LogsFile() io.WriteSyncer
	}

	factory struct {
		sync.Mutex
		runners map[string]func(ctx Context, payload map[string]interface{}) (Runner, error)
	}
)

var (
	runners factory
	once    sync.Once
)

// GenerateID creates a new ID.
func GenerateID() string {
	time.Sleep(time.Nanosecond)
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// Register registers a runner constructor.
func Register(name string, fn func(ctx Context, payload map[string]interface{}) (Runner, error)) {
	once.Do(func() {
		runners = factory{
			runners: make(map[string]func(ctx Context, payload map[string]interface{}) (Runner, error)),
		}
	})

	runners.Lock()
	defer runners.Unlock()

	runners.runners[name] = fn
}

// Lookup returns a new runner according given payload.
func Lookup(ctx Context, payload map[string]interface{}) (Runner, error) {
	runners.Lock()
	defer runners.Unlock()

	for k := range payload {
		if create, ok := runners.runners[k]; ok {
			return create(ctx, payload)
		}
	}

	return nil, errors.New("runner not found")
}
