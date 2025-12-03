package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var profile *string

var rootCmd = &cobra.Command{
	Use:   "dx",
	Short: "Developer experience tool for Kubernetes local development",
	Long: `DX streamlines local Kubernetes development by managing Docker builds,
Helm deployments, and development proxies.

Configuration is stored in ~/.dx-config.yaml. Run 'dx initialize' to create
a sample configuration file.

Common workflows:
  dx build                    Build all Docker images for the default profile
  dx install                  Deploy all services to local Kubernetes
  dx update                   Build and reinstall services
  dx context set <name>       Switch to a different context`,
}

func Execute() {
	profile = rootCmd.PersistentFlags().StringP("profile", "p", DefaultProfile, "Profile to use")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
