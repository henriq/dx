package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var version = "dev"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays the application version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("DX %s\n", version)
	},
}
