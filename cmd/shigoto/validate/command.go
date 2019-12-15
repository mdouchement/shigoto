package validate

import (
	"fmt"

	"github.com/mdouchement/shigoto/pkg/shigoto"
	"github.com/spf13/cobra"
)

// Command launches the validate subcommand.
var Command = &cobra.Command{
	Use:   "validate",
	Short: "Validate given shigoto file",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		if _, err := shigoto.Load(args[0]); err != nil {
			return err
		}
		fmt.Println("OK")
		return nil
	},
}
