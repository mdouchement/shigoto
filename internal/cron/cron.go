package cron

import (
	"sync"

	"github.com/mdouchement/logger"
	"github.com/mdouchement/shigoto/pkg/runner"
	"github.com/mdouchement/shigoto/pkg/shigoto"
	"github.com/robfig/cron/v3"
)

// A Pool contains a pool of crons and carries lots of helping methods.
type Pool struct {
	mu      sync.Mutex
	logger  logger.Logger
	running map[string]bool
	cron    map[string]*cron.Cron
	shigoto map[string]*shigoto.Shigoto
}

// New returns a new Pool.
func New(l logger.Logger) *Pool {
	return &Pool{
		logger:  l,
		running: make(map[string]bool),
		cron:    make(map[string]*cron.Cron),
		shigoto: make(map[string]*shigoto.Shigoto),
	}
}

// Register adds the given shigoto to the cron pool.
func (p *Pool) Register(s *shigoto.Shigoto) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.shigoto[s.Name] = s
	cron := cron.New(cron.WithLogger(
		cron.PrintfLogger(p.logger)),
		cron.WithChain(
			cron.SkipIfStillRunning(cron.PrintfLogger(p.logger)),
		),
	)
	p.cron[s.Name] = cron
	for _, baito := range s.Baito {
		chain := runner.Chain(baito.Commands()...)
		chain.AttachLogger(p.logger)
		cron.Schedule(baito.Schedule(), chain)

		p.logger.Infof(`New job registered "%s" - "%s"`, baito.Name(), baito.Schedule())
	}
}

// Get returns the registred shigoto according the given name.
func (p *Pool) Get(name string) *shigoto.Shigoto {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.shigoto[name]
}

// Start start all schedulers.
func (p *Pool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for name, cron := range p.cron {
		if p.running[name] {
			continue
		}

		p.logger.Infof("Starting scheduler '%s' with %d jobs", name, len(p.shigoto[name].Baito))
		cron.Start()
		p.running[name] = true
	}
}

// Stop stops and waits for the termination of the tasks of all schedulers.
func (p *Pool) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for name, cron := range p.cron {
		p.logger.Infof("Shuting down the scheduler '%s'...", name)
		<-cron.Stop().Done() // Wait for the termination of the tasks.

		delete(p.running, name)
	}
}

// StartShigoto starts the scheduler of the given shigoto's name.
func (p *Pool) StartShigoto(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running[name] {
		return
	}

	p.logger.Infof("Starting scheduler '%s' with %d jobs", name, len(p.shigoto[name].Baito))
	p.cron[name].Start()
	p.running[name] = true
}

// StopShigoto stops and waits for the termination of the tasks of the given shigoto.
func (p *Pool) StopShigoto(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.logger.Infof("Shuting down the scheduler '%s'...", name)
	<-p.cron[name].Stop().Done() // Wait for the termination of the tasks.

	delete(p.shigoto, name)
	delete(p.cron, name)
	delete(p.running, name)
}
