package shigoto

import (
	"fmt"
	"os"

	"github.com/knadh/koanf"
	"github.com/mdouchement/shigoto/pkg/io"
	"github.com/mdouchement/shigoto/pkg/runner"
	"github.com/mdouchement/shigoto/pkg/templater"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

type (
	// A Baito (aka Arubaito from German Arbeit) is a task/job that will be runned at a scheduled time.
	Baito struct {
		FieldName        string
		FieldSchedule    Schedule
		FieldWorkdir     string
		FieldLogsFile    io.WriteSyncer
		FieldVariables   map[string]string
		FieldEnvironment map[string]string
		FieldCommands    []runner.Runner
	}

	// A Schedule describes a job's duty cycle.
	Schedule cron.Schedule

	schedule struct {
		Schedule
		raw string
	}
)

func (s *schedule) String() string {
	return s.raw
}

// Name returns the name.
func (b *Baito) Name() string {
	return b.FieldName
}

// Schedule returns the schedule.
func (b *Baito) Schedule() Schedule {
	return b.FieldSchedule
}

// Workdir returns the working directory.
func (b *Baito) Workdir() string {
	return b.FieldWorkdir
}

// LogsFile returns the logs file where stdout/stderr are redirected.
func (b *Baito) LogsFile() io.WriteSyncer {
	return b.FieldLogsFile
}

// Variables returns the variables.
func (b *Baito) Variables() map[string]string {
	return b.FieldVariables
}

// Environment returns the environment variables.
func (b *Baito) Environment() map[string]string {
	return b.FieldEnvironment
}

// Commands returns the commands to be executed.
func (b *Baito) Commands() []runner.Runner {
	return b.FieldCommands
}

// ExpandEnv replaces ${var} or $var in the string according to the values
// of the current environment variables. References to undefined
// variables left as there are.
func (b *Baito) ExpandEnv(str string) string {
	str = os.Expand(str, func(k string) string {
		if e, ok := b.FieldEnvironment[k]; ok {
			return e
		}

		return fmt.Sprintf("${%s}", k)
	})
	return os.ExpandEnv(str)
}

// ExpandVariables replace templatized variables by their values.
func (b *Baito) ExpandVariables(str string) string {
	return templater.New(b).Replace(str)
}

// ExpandAll replace templatized variables and environment variables by their values.
func (b *Baito) ExpandAll(str string) string {
	s := b.ExpandVariables(str)
	return b.ExpandEnv(s)
}

func loadBaito(konf *koanf.Koanf, name string) (*Baito, error) {
	baito := &Baito{
		FieldName: name,
	}

	baito.loadVariables(konf)
	baito.loadEnvironment(konf)
	baito.loadWorkdir(konf)

	if err := baito.loadSchedule(konf); err != nil {
		return nil, err
	}

	if err := baito.loadLogsFile(konf); err != nil {
		return nil, err
	}

	if err := baito.loadCommands(konf); err != nil {
		return nil, err
	}

	return baito, nil
}

func (b *Baito) loadVariables(konf *koanf.Koanf) {
	b.FieldVariables = konf.StringMap(globalvariables)

	path := fmt.Sprintf("%s.%s.variables", entrypoint, b.FieldName)
	if !konf.Exists(path) {
		return
	}

	for k, v := range konf.StringMap(path) {
		b.FieldVariables[k] = b.ExpandVariables(v)
	}
}

func (b *Baito) loadEnvironment(konf *koanf.Koanf) {
	b.FieldEnvironment = konf.StringMap(globalenv)

	path := fmt.Sprintf("%s.%s.environment", entrypoint, b.FieldName)
	if !konf.Exists(path) {
		return
	}

	for k, v := range konf.StringMap(path) {
		b.FieldEnvironment[k] = b.ExpandAll(v)
	}
}

func (b *Baito) loadWorkdir(konf *koanf.Koanf) {
	path := fmt.Sprintf("%s.%s.workdir", entrypoint, b.FieldName)
	b.FieldWorkdir = b.ExpandAll(konf.String(path))
}

func (b *Baito) loadSchedule(konf *koanf.Koanf) error {
	path := fmt.Sprintf("%s.%s.schedule", entrypoint, b.FieldName)
	if !konf.Exists(path) {
		return errors.Errorf("%s: missing schedule", path)
	}

	s := &schedule{
		raw: konf.String(path),
	}
	b.FieldSchedule = s

	var err error
	s.Schedule, err = cron.ParseStandard(s.raw)
	return errors.Wrap(err, path)
}

func (b *Baito) loadLogsFile(konf *koanf.Koanf) (err error) {
	path := fmt.Sprintf("%s.%s.logs_file", entrypoint, b.FieldName)

	logfile := konf.String(path)
	if logfile == "" {
		return
	}
	logfile = b.ExpandAll(logfile)

	b.FieldLogsFile, err = os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	return errors.Wrap(err, "could not create logs redirection file")
}

func (b *Baito) loadCommands(konf *koanf.Koanf) error {
	path := fmt.Sprintf("%s.%s.commands", entrypoint, b.FieldName)
	if !konf.Exists(path) {
		return errors.Errorf("%s: missing commands", path)
	}

	sl, ok := konf.Get(path).([]interface{})
	if !ok {
		return errors.Errorf("%s: expected commands to be an array", path)
	}

	templater := templater.New(b)
	for i, command := range sl {
		var err error
		var c runner.Runner

		switch v := command.(type) {
		case string:
			v = templater.Replace(v)
			c, err = runner.Lookup(b, map[string]interface{}{"exec": v})
		case map[string]interface{}:
			v = templater.ReplaceMapI(v)
			c, err = runner.Lookup(b, v)
		default:
			return errors.Errorf("%s: invalid command format", path)
		}

		if err != nil {
			return errors.Wrapf(err, "%s[%d]: load", path, i)
		}
		b.FieldCommands = append(b.FieldCommands, c)
	}

	return templater.Err()
}
