package cmd

import (
	"fmt"

	"dx/cmd/cli/app"
	"github.com/spf13/cobra"
)

func init() {
	secretCmd.AddCommand(secretsListCmd)
	secretCmd.AddCommand(secretDeleteCmd)
	secretCmd.AddCommand(secretSetCmd)
	rootCmd.AddCommand(secretCmd)
}

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manages secrets for the application",
	Long:  `Commands for managing and viewing secrets using an application specific encrypted storage`,
}

var secretSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Sets a secret",
	Long:  `Saves a secret to the application's encrypted storage'`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		handler, err := app.InjectSecretCommandHandler()
		if err != nil {
			return err
		}

		return handler.HandleSet(args[0], args[1])
	},
}

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all secrets",
	Long:  `Lists all secrets in the application`,
	RunE: func(cmd *cobra.Command, args []string) error {
		handler, err := app.InjectSecretCommandHandler()
		if err != nil {
			return err
		}

		return handler.HandleList()
	},
}

var secretDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a secret",
	Long:  `Deletes a secret from the application's encrypted storage`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		secretsRepo, err := app.InjectSecretRepository()
		if err != nil {
			return err
		}
		configRepo, err := app.InjectConfigRepo()
		if err != nil {
			return fmt.Errorf("error injecting config repo: %v", err)
		}
		configContext, err := configRepo.LoadCurrentConfigurationContext()
		if err != nil {
			return fmt.Errorf("error loading current configuration context: %v", err)
		}

		secrets, err := secretsRepo.LoadSecrets(configContext.Name)
		if err != nil {
			return err
		}

		for _, secret := range secrets {
			if secret.Key == args[0] {
				return nil
			}
		}
		return fmt.Errorf("secret '%s' not found", args[0])
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		secretsRepo, err := app.InjectSecretRepository()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		configRepo, err := app.InjectConfigRepo()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		currentContextName, err := configRepo.LoadCurrentContextName()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		secrets, err := secretsRepo.LoadSecrets(currentContextName)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		var secretKeys []string
		for _, secret := range secrets {
			secretKeys = append(secretKeys, secret.Key)
		}
		return secretKeys, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		handler, err := app.InjectSecretCommandHandler()
		if err != nil {
			return err
		}

		return handler.HandleDelete(args[0])
	},
}
