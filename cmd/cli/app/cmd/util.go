package cmd

import (
	"fmt"

	"dx/cmd/cli/app"
	"github.com/spf13/cobra"
)

const DefaultProfile = "default"

func ServiceArgsValidator(cmd *cobra.Command, args []string) error {
	configRepo, err := app.InjectConfigRepo()
	if err != nil {
		return fmt.Errorf("error injecting config repo: %v", err)
	}
	configContext, err := configRepo.LoadCurrentConfigurationContext()
	if err != nil {
		return fmt.Errorf("error loading current configuration context: %v", err)
	}
	for _, service := range args {
		foundService := false
		for _, s := range configContext.Services {
			if service == s.Name {
				foundService = true
				break
			}
		}
		if !foundService {
			return fmt.Errorf("service %s not found", service)
		}
	}

	return nil
}

func ServiceArgsCompletion(
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]cobra.Completion, cobra.ShellCompDirective) {
	configRepo, err := app.InjectConfigRepo()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	configContext, err := configRepo.LoadCurrentConfigurationContext()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var matchingServices []string
	for _, s := range configContext.Services {
		matchingServices = append(matchingServices, s.Name)
	}

	return matchingServices, cobra.ShellCompDirectiveNoFileComp
}
