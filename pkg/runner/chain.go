package runner

type chain struct {
	base

	runners []Runner
}

// Chain wraps and runs sequentially the given runners.
func Chain(runners ...Runner) Runner {
	return &chain{runners: runners}
}

func (r *chain) Run() {
	r.Lock()
	defer r.Unlock()

	// Reset previous state
	r.err = nil

	//

	for _, runner := range r.runners {
		runner.AttachLogger(r.log.WithField("chain", GenerateID()))

		if runner.IsDeferrable() {
			defer func(runner Runner) {
				runner.Run()
				if !runner.IsErrorIgnored() && runner.Error() != nil {
					r.err = runner.Error()
				}
			}(runner)

			continue
		}

		runner.Run()
		if !runner.IsErrorIgnored() && runner.Error() != nil {
			r.err = runner.Error()
			return
		}
	}
}
