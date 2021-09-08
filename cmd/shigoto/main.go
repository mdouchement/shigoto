package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/mdouchement/shigoto/cmd/shigoto/daemon"
	"github.com/mdouchement/shigoto/cmd/shigoto/reload"
	"github.com/mdouchement/shigoto/cmd/shigoto/validate"
	"github.com/spf13/cobra"
)

var (
	version  = "dev"
	revision = "none"
	date     = "unknown"
)

func main() {
	c := &cobra.Command{
		Use:     "shigoto",
		Short:   "Shigoto, a nextgen crontab",
		Version: fmt.Sprintf("%s - build %.7s @ %s - %s", version, revision, date, runtime.Version()),
		Args:    cobra.NoArgs,
	}
	c.AddCommand(daemon.Command)
	c.AddCommand(reload.Command)
	c.AddCommand(validate.Command)
	c.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Version for shigoto",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(c.Version)
		},
	})

	if err := c.Execute(); err != nil {
		log.Fatal(err)
	}
}
