package cmd

import (
	"dx/cmd/cli/app"

	"github.com/spf13/cobra"
)

var skipDevProxy *bool

func init() {
	skipDevProxy = installCmd.Flags().BoolP("skip-dev-proxy", "s", false, "Skip dev proxy installation")
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install [service...]",
	Short: "Deploy services to Kubernetes via Helm",
	Long: `Deploys the specified services to the local Kubernetes cluster using Helm.
If no services are specified, deploys all services in the current profile.

This command also sets up the dev-proxy for routing traffic between local
and Kubernetes services (unless --skip-dev-proxy is specified).`,
	Example: `  # Install all services in the default profile
  dx install

  # Install specific services
  dx install api database

  # Install without dev-proxy setup
  dx install --skip-dev-proxy

  # Install all services regardless of profile
  dx install -p all`,
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
