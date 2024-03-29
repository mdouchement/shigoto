package daemon

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/mdouchement/logger"
	"github.com/mdouchement/shigoto/internal/config"
	"github.com/mdouchement/shigoto/internal/cron"
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
		Use:   "daemon",
		Short: "Start Shigoto service",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			if cfg == "" {
				var err error
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

			l := slog.New(logger.NewSlogTextHandler(os.Stdout, &logger.SlogTextOption{
				Level:           slog.LevelInfo,
				DisableColors:   !konf.Bool("log.force_color"),
				ForceColors:     konf.Bool("log.force_color"),
				ForceFormatting: konf.Bool("log.force_formating"),
				PrefixRE:        regexp.MustCompile(`^(\[.*?\])\s`),
				FullTimestamp:   true,
				TimestampFormat: "2006-01-02 15:04:05",
			}))
			log := logger.WrapSlog(l)
			pool := cron.New(log)

			//
			//

			sock := socket.New(konf.String("socket"))
			defer sock.Close()

			go func() {
				err := sock.Listen(func(event []byte) []byte {
					if bytes.Equal(event, socket.SignalReload) {
						log.Info("Reloading daemon")

						err := cron.Load(filepath.Join(konf.String("directory")), pool, log)
						if err != nil {
							log.WithError(err).Error("Fail to reloading")
							return []byte(err.Error())
						}

						log.Info("Reloaded")
						return []byte("OK")
					}
					return []byte("Unsupported signal")
				})
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}()

			//
			//

			err := cron.Load(filepath.Join(konf.String("directory")), pool, log)
			if err != nil {
				return err
			}
			pool.Start()
			defer pool.Stop()

			signals := make(chan os.Signal, 1)
			signal.Notify(signals, os.Interrupt, os.Kill)
			<-signals

			return nil
		},
	}

	cfg string
)
