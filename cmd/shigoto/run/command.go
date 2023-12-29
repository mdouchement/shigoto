package run

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/mdouchement/logger"
	"github.com/mdouchement/shigoto/pkg/runner"
	"github.com/mdouchement/shigoto/pkg/shigoto"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Command launches the validate subcommand.
var Command = &cobra.Command{
	Use:   "run file.yml ALL|baito [baito]...",
	Short: "Run (oneshot) the given shigoto file",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		logrus := logrus.New()
		logrus.SetFormatter(&logger.LogrusTextFormatter{
			ForceColors:     true,
			ForceFormatting: true,
			PrefixRE:        regexp.MustCompile(`^(\[.*?\])\s`),
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
		log := logger.WrapLogrus(logrus)

		//

		shigoto, err := shigoto.Load(args[0])
		if err != nil {
			return err
		}

		var selected []runner.Runner
		for _, baito := range shigoto.Baito {
			fmt.Println("Found baito:", baito.Name())

			if len(args) <= 1 {
				continue
			}

			if args[1] == "ALL" {
				selected = append(selected, baito.Commands()...)
				continue
			}

			if slices.Contains[[]string, string](args[1:], baito.Name()) {
				selected = append(selected, baito.Commands()...)
			}
		}
		fmt.Println("---")

		chain := runner.Chain(selected...)
		chain.AttachLogger(log)
		chain.Run()
		return chain.Error()
	},
}
