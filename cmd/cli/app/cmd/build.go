package cmd

import (
	"dx/cmd/cli/app"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:               "build [service...]",
	Short:             "Builds the application",
	Long:              `Builds the selected services if arguments are supplied, otherwise builds all services`,
	Args:              ServiceArgsValidator,
	ValidArgsFunction: ServiceArgsCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		handler, err := app.InjectBuildCommandHandler()
		if err != nil {
			return err
		}

		return handler.Handle(args, *profile)
	},
}
