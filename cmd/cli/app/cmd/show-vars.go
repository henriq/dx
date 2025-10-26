package cmd

import (
	"dx/cmd/cli/app"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(showVarsCommand)
}

var showVarsCommand = &cobra.Command{
	Use:   "show-vars",
	Short: "Shows variables that can be used in templates.",
	Long:  `Shows all variables that can be used in templates`,
	RunE: func(cmd *cobra.Command, args []string) error {
		handler, err := app.InjectShowVarsCommandHandler()
		if err != nil {
			return err
		}

		return handler.Handle()
	},
}
