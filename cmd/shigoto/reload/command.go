package reload

import (
	"fmt"
	"os"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/mdouchement/shigoto/internal/config"
	"github.com/mdouchement/shigoto/internal/socket"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	Command.Flags().StringVarP(&cfg, "config", "c", "", "Configuration file")
}

var (
	// Command launches the deamon subcommand.
	Command = &cobra.Command{
		Use:   "reload",
		Short: "Realod Shigoto service",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) (err error) {
			if cfg == "" {
				cfg, err = config.Lookup(config.Filenames...)
				if err != nil {
					if err == os.ErrNotExist {
						return errors.New("no configuration found from the default pathes")
					}
					return err
				}
			}

			konf := koanf.New(".")
			if err := konf.Load(file.Provider(cfg), toml.Parser()); err != nil {
				return err
			}

			payload, err := socket.New(konf.String("socket")).Request(socket.SignalReload)
			if err != nil {
				return err
			}
			fmt.Println(string(payload))
			return nil
		},
	}

	cfg string
)
