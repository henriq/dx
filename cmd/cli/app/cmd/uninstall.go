package cmd

import (
	"dx/cmd/cli/app"

	"github.com/spf13/cobra"
)

var skipDevProxy *bool

func init() {
	skipDevProxy = installCmd.Flags().BoolP("skip-dev-proxy,", "s", false, "Skip dev proxy installation")
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:               "install [service...]",
	Short:             "Installs the application",
	Long:              `Installs the selected services if arguments are supplied, otherwise installs all services`,
	Args:              ServiceArgsValidator,
	ValidArgsFunction: ServiceArgsCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		handler, err := app.InjectInstallCommandHandler()
		if err != nil {
			return err
		}

		return handler.Handle(args, *profile, *skipDevProxy)
	},
}
