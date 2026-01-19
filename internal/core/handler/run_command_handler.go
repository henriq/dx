package handler

import (
	"fmt"
	"regexp"
	"strings"

	"dx/internal/cli/output"
	"dx/internal/core"
	"dx/internal/core/domain"
	"dx/internal/ports"
)

type RunCommandHandler struct {
	configRepository  core.ConfigRepository
	secretsRepository core.SecretsRepository
	templater         ports.Templater
	scm               ports.Scm
	commandRunner     ports.CommandRunner
}

func ProvideRunCommandHandler(
	configRepository core.ConfigRepository,
	secretsRepository core.SecretsRepository,
	templater ports.Templater,
	scm ports.Scm,
	commandRunner ports.CommandRunner,
) RunCommandHandler {
	return RunCommandHandler{
		configRepository:  configRepository,
		secretsRepository: secretsRepository,
		templater:         templater,
		scm:               scm,
		commandRunner:     commandRunner,
	}
}

func (h *RunCommandHandler) Handle(scripts map[string]string, executionPlan []string) error {
	renderValues, err := core.CreateTemplatingValues(h.configRepository, h.secretsRepository)
	if err != nil {
		return err
	}

	configContext, err := h.configRepository.LoadCurrentConfigurationContext()
	if err != nil {
		return err
	}

	for _, scriptName := range executionPlan {
		script := scripts[scriptName]

		dependentServices, err := findServiceDependencies(script, configContext.Services)
		if err != nil {
			return err
		}

		if len(dependentServices) > 0 {
			var dependentServiceNames []string
			for _, service := range dependentServices {
				dependentServiceNames = append(dependentServiceNames, service.Name)
			}
			output.PrintStep(fmt.Sprintf("Updating repositories: %s", output.Dim(strings.Join(dependentServiceNames, ", "))))
		}

		for _, dependentService := range dependentServices {
			if dependentService.GitRepoPath == "" || dependentService.GitRef == "" {
				return fmt.Errorf("git repository path or ref is empty for service '%s'", dependentService.Name)
			}
			err = h.scm.Download(dependentService.GitRepoPath, dependentService.GitRef, dependentService.Path)
			if err != nil {
				return err
			}
		}

		renderedScript, err := h.templater.Render(script, scriptName, renderValues)
		if err != nil {
			return err
		}

		output.PrintStep(fmt.Sprintf("Running %s", output.Bold(scriptName)))
		fmt.Println()

		if err := h.commandRunner.RunInteractive("bash", "-c", renderedScript); err != nil {
			fmt.Println()
			return fmt.Errorf("script '%s' failed: %v", scriptName, err)
		}

		fmt.Println()
	}
	return nil
}

func findServiceDependencies(script string, existingServices []domain.Service) ([]domain.Service, error) {
	serviceRegex := regexp.MustCompile(`{{[^}]*\.Services\s*\.*"*([\w\-_]+)[^}]*path`)
	matches := serviceRegex.FindAllStringSubmatch(script, -1)

	services := make([]domain.Service, 0)
	for _, match := range matches {
		if len(match) > 1 {
			service, err := findService(match[1], existingServices)
			if err != nil {
				return nil, err
			}
			services = append(services, service)
		}
	}

	return services, nil
}

func findService(serviceName string, existingServices []domain.Service) (domain.Service, error) {
	for _, service := range existingServices {
		if service.Name == serviceName {
			return service, nil
		}
	}
	return domain.Service{}, fmt.Errorf("service '%s' not found", serviceName)
}
