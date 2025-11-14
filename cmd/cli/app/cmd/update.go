package cmd

import (
	"dx/cmd/cli/app"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update [service...]",
	Short: "Updates the application",
	Long: `Builds and reinstalls the selected services if arguments are supplied,
otherwise builds and reinstalls all services`,
	Args:              ServiceArgsValidator,
	ValidArgsFunction: ServiceArgsCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		buildHandler, err := app.InjectBuildCommandHandler()
		installHandler, err := app.InjectInstallCommandHandler()
		if err != nil {
			return err
		}

		err = buildHandler.Handle(args, *profile)
		if err != nil {
			return err
		}

		return installHandler.Handle(args, *profile, false)
	},
}
