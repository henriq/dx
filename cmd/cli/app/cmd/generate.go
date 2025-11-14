package cmd

import (
	"dx/cmd/cli/app"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	generateCmd.AddCommand(generateHostEntriesCmd)
	rootCmd.AddCommand(generateCmd)
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates data based on configuration",
	Long:  `Commands for generating data based on configuration`,
}

var generateHostEntriesCmd = &cobra.Command{
	Use:   "host-entries",
	Short: "Generates host entries for putting in the host file based on all configuration contexts",
	Long:  `Generates host entries in a format that can be inserted in the hosts file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		handler, err := app.InjectGenerateCommandHandler()
		if err != nil {
			return err
		}

		return handler.HandleGenerateHostEntries(os.Stdout)
	},
}
