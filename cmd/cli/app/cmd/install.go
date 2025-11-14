package cmd

import (
	"dx/cmd/cli/app"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

var uninstallCmd = &cobra.Command{
	Use:               "uninstall [service...]",
	Short:             "Uninstalls the application",
	Long:              `Uninstalls the selected services if arguments are supplied, otherwise installs all services`,
	Args:              ServiceArgsValidator,
	ValidArgsFunction: ServiceArgsCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		handler, err := app.InjectUninstallCommandHandler()
		if err != nil {
			return err
		}

		return handler.Handle(args, *profile)
	},
}
