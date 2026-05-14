package cmd

import (
	"pilot/cmd/cli/app"

	"github.com/spf13/cobra"
)

var buildImageSourceOverrides []string

func init() {
	rootCmd.AddCommand(buildCmd)
	RegisterImageSourceOverrideFlag(buildCmd, &buildImageSourceOverrides)
}

var buildCmd = &cobra.Command{
	Use:   "build [service...]",
	Short: "Build Docker images for services",
	Long: `Builds Docker images for the specified services. If no services are
specified, builds all services in the current profile.

Images are built using the configured Dockerfile and made available to the
local Kubernetes cluster.

Use --image-source to build an image from a local directory instead of
cloning its configured git repo — handy for testing local changes before
committing or pushing.`,
	Example: `  # Build all services in the default profile
  pilot build

  # Build specific services
  pilot build api frontend

  # Build all services regardless of profile
  pilot build -p all

  # Build an image from a local checkout
  pilot build api --image-source api:./services/api`,
	Args:              ServiceArgsValidator,
	ValidArgsFunction: ServiceArgsCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		imageSourceOverrides, err := ParseImageSourceOverrides(buildImageSourceOverrides)
		if err != nil {
			return err
		}
		handler, err := app.InjectBuildCommandHandler()
		if err != nil {
			return err
		}

		return handler.Handle(args, *profile, imageSourceOverrides)
	},
}
