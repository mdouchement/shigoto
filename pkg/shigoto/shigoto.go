package shigoto

import (
	"path/filepath"
	"reflect"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
)

const (
	globalvariables = "variables"
	globalenv       = "environment"
	entrypoint      = "shigoto"
)

// Shigoto represents a shigoto.yml
type Shigoto struct {
	Name  string
	Baito map[string]Baito
	konf  *koanf.Koanf
}

// Load loads a Shigoto from the given filename.
func Load(filename string) (*Shigoto, error) {
	konf := koanf.New(".")
	if err := konf.Load(file.Provider(filename), yaml.Parser()); err != nil {
		return nil, err
	}

	shigoto := &Shigoto{
		Name:  filepath.Base(filename),
		Baito: map[string]Baito{},
		konf:  konf,
	}

	for _, name := range konf.MapKeys(entrypoint) {
		b, err := loadBaito(konf, name)
		if err != nil {
			return nil, err
		}

		shigoto.Baito[name] = *b
	}

	return shigoto, nil
}

// Same returns true if both shigoto are the same.
func (s *Shigoto) Same(shigoto *Shigoto) bool {
	return reflect.DeepEqual(s.konf, shigoto.konf)
}
