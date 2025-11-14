package cmd

import (
	"dx/cmd/cli/app"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initializeCmd)
}

var initializeCmd = &cobra.Command{
	Use:   "initialize",
	Short: "Generates a new configuration file with sample values",
	Long:  `A new configuration file is written to ~/.dx-config.yaml. This file contains sample values for all configuration options. The file is not created if it already exists.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		handler, err := app.InjectInitializeCommandHandler()
		if err != nil {
			return err
		}

		return handler.Handle()
	},
}
