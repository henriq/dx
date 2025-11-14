package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var profile *string

var rootCmd = &cobra.Command{
	Use:   "dx",
	Short: "DX is a tool for managing a local development environment using kubernetes.",
	Long:  `DX is a tool for managing a local development environment using kubernetes.`,
}

func Execute() {
	profile = rootCmd.PersistentFlags().StringP("profile", "p", DefaultProfile, "Profile to use")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
