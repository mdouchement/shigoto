package config

import "os"

// Filenames is the default configuration file pathes.
var Filenames = []string{
	"/etc/shigoto/shigoto.toml",
	"/shigoto/shigoto.toml",
	"./shigoto.toml",
}

// Lookup returns the first found filename.
func Lookup(filenames ...string) (string, error) {
	for _, filename := range filenames {
		_, err := os.Stat(filename)
		if err == nil {
			return filename, nil
		}
		if os.IsNotExist(err) {
			continue
		}

		return "", err
	}

	return "", os.ErrNotExist
}
