package cmd

import (
	"dx/cmd/cli/app"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(genEnvKeyCmd)
}

var genEnvKeyCmd = &cobra.Command{
	Use:   "gen-env-key",
	Short: "Generates an environment key for the currently active cluster configuration",
	Long:  `Generates an environment key using the current cluster and namespace configuration in ~/.kube/config`,
	RunE: func(cmd *cobra.Command, args []string) error {
		handler, err := app.InjectGenEnvKeyCommandHandler()
		if err != nil {
			return err
		}

		return handler.Handle()
	},
}
