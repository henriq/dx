package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pilot/cmd/cli/app"
	"pilot/internal/core/domain"

	"github.com/spf13/cobra"
)

const DefaultProfile = "default"

func ServiceArgsValidator(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	if home, err := os.UserHomeDir(); err == nil {
		if _, err := os.Stat(filepath.Join(home, ".pilot-config.yaml")); err != nil {
			return nil
		}
	}
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
	var services []string
	for _, s := range configContext.Services {
		services = append(services, s.Name)
	}

	return services, cobra.ShellCompDirectiveNoFileComp
}

func SecretKeysCompletion(
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
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
}

// RegisterHelmChartOverrideFlag wires the --helm-chart flag onto cmd, binding
// repeated "service:path" values into target.
//
// The delimiter is ":" rather than "=" to avoid a misfire in Cobra's zsh and
// fish completion scripts, whose `-.*=` regex mangles values containing a dash.
func RegisterHelmChartOverrideFlag(cmd *cobra.Command, target *[]string) {
	cmd.Flags().StringArrayVar(
		target,
		"helm-chart",
		nil,
		"render a service's Helm chart from a local directory instead of cloning the repo (useful for chart development); format: service:path; repeat for multiple services",
	)
	_ = cmd.RegisterFlagCompletionFunc("helm-chart", HelmChartOverrideCompletion)
}

// HelmChartOverrideCompletion offers service names with a trailing ":" until
// the user types one, then switches to directory completion for the path.
func HelmChartOverrideCompletion(
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]cobra.Completion, cobra.ShellCompDirective) {
	if separatorIndex := strings.Index(toComplete, ":"); separatorIndex >= 0 {
		servicePrefix := toComplete[:separatorIndex+1]
		pathToComplete := toComplete[separatorIndex+1:]
		return helmChartOverrideDirectoryCompletions(servicePrefix, pathToComplete), cobra.ShellCompDirectiveNoSpace
	}
	configRepo, err := app.InjectConfigRepo()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	configContext, err := configRepo.LoadCurrentConfigurationContext()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return helmChartOverrideServiceCompletions(configContext.Services), cobra.ShellCompDirectiveNoSpace
}

func expandLeadingTilde(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") && !strings.HasPrefix(path, `~\`) {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if path == "~" {
		return home, nil
	}
	expanded := filepath.Join(home, path[2:])
	if strings.HasSuffix(path, "/") || strings.HasSuffix(path, `\`) {
		expanded += string(filepath.Separator)
	}
	return expanded, nil
}

// On Windows a bare "C:" means "current directory on drive C:" rather than the
// drive root; appending the separator forces filepath.Split to treat it as the root.
func anchorBareVolumeToRoot(path, volumeName, pathSeparator string) string {
	if volumeName != "" && path == volumeName {
		return path + pathSeparator
	}
	return path
}

func helmChartOverrideDirectoryCompletions(servicePrefix, pathToComplete string) []cobra.Completion {
	if pathToComplete == "~" {
		pathToComplete = "~" + string(filepath.Separator)
	}
	pathToComplete = anchorBareVolumeToRoot(
		pathToComplete,
		filepath.VolumeName(pathToComplete),
		string(filepath.Separator),
	)
	directoryToList, fileNamePrefix := filepath.Split(pathToComplete)
	listTarget, err := expandLeadingTilde(directoryToList)
	if err != nil {
		return nil
	}
	if listTarget == "" {
		listTarget = "."
	}
	entries, err := os.ReadDir(listTarget)
	if err != nil {
		return nil
	}
	completions := make([]cobra.Completion, 0, len(entries))
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), fileNamePrefix) {
			continue
		}
		if strings.HasPrefix(entry.Name(), ".") && !strings.HasPrefix(fileNamePrefix, ".") {
			continue
		}
		entryInfo, err := os.Stat(filepath.Join(listTarget, entry.Name()))
		if err != nil || !entryInfo.IsDir() {
			continue
		}
		completions = append(completions, servicePrefix+directoryToList+entry.Name()+string(filepath.Separator))
	}
	return completions
}

func helmChartOverrideServiceCompletions(services []domain.Service) []cobra.Completion {
	completions := make([]cobra.Completion, 0, len(services))
	for _, service := range services {
		completions = append(completions, service.Name+":")
	}
	return completions
}

// RegisterImageSourceOverrideFlag wires the --image-source flag onto cmd,
// binding repeated "image:path" values into target.
//
// The delimiter is ":" rather than "=" to avoid a misfire in Cobra's zsh and
// fish completion scripts, whose `-.*=` regex mangles values containing a dash.
func RegisterImageSourceOverrideFlag(cmd *cobra.Command, target *[]string) {
	cmd.Flags().StringArrayVar(
		target,
		"image-source",
		nil,
		"build a Docker image from a local directory instead of cloning the repo (useful for testing local image changes); format: image:path; repeat for multiple images",
	)
	_ = cmd.RegisterFlagCompletionFunc("image-source", ImageSourceOverrideCompletion)
}

// ImageSourceOverrideCompletion offers Docker image names with a trailing ":"
// until the user types one, then switches to directory completion for the path.
func ImageSourceOverrideCompletion(
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]cobra.Completion, cobra.ShellCompDirective) {
	if separatorIndex := strings.Index(toComplete, ":"); separatorIndex >= 0 {
		imagePrefix := toComplete[:separatorIndex+1]
		pathToComplete := toComplete[separatorIndex+1:]
		return helmChartOverrideDirectoryCompletions(imagePrefix, pathToComplete), cobra.ShellCompDirectiveNoSpace
	}
	configRepo, err := app.InjectConfigRepo()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	configContext, err := configRepo.LoadCurrentConfigurationContext()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return imageSourceOverrideImageCompletions(configContext.Services), cobra.ShellCompDirectiveNoSpace
}

func imageSourceOverrideImageCompletions(services []domain.Service) []cobra.Completion {
	seen := map[string]struct{}{}
	var completions []cobra.Completion
	for _, service := range services {
		for _, image := range service.DockerImages {
			if _, ok := seen[image.Name]; ok {
				continue
			}
			seen[image.Name] = struct{}{}
			completions = append(completions, image.Name+":")
		}
	}
	return completions
}

// ParseImageSourceOverrides converts repeated --image-source entries in
// "image:path" form into a DockerImageSourceOverrides value. Duplicate image
// entries take the last value.
func ParseImageSourceOverrides(entries []string) (domain.DockerImageSourceOverrides, error) {
	sourcePathByImage := map[string]string{}
	for _, entry := range entries {
		imageName, sourcePath, found := strings.Cut(entry, ":")
		if !found {
			return domain.DockerImageSourceOverrides{}, fmt.Errorf(
				"invalid --image-source value %q; expected format: image:path",
				entry,
			)
		}
		imageName = strings.TrimSpace(imageName)
		sourcePath = strings.TrimSpace(sourcePath)
		if imageName == "" {
			return domain.DockerImageSourceOverrides{}, fmt.Errorf(
				"invalid --image-source value %q; image name cannot be empty",
				entry,
			)
		}
		if sourcePath == "" {
			return domain.DockerImageSourceOverrides{}, fmt.Errorf(
				"invalid --image-source value %q; source path cannot be empty",
				entry,
			)
		}
		if index := strings.IndexFunc(sourcePath, func(r rune) bool {
			return r < 0x20 || r == 0x7f
		}); index >= 0 {
			return domain.DockerImageSourceOverrides{}, fmt.Errorf(
				"invalid --image-source value %q; source path contains a control character at byte %d",
				entry,
				index,
			)
		}
		expandedSourcePath, err := expandLeadingTilde(sourcePath)
		if err != nil {
			return domain.DockerImageSourceOverrides{}, fmt.Errorf(
				"failed to resolve --image-source path %q: %w",
				sourcePath,
				err,
			)
		}
		absoluteSourcePath, err := filepath.Abs(expandedSourcePath)
		if err != nil {
			return domain.DockerImageSourceOverrides{}, fmt.Errorf(
				"failed to resolve --image-source path %q: %w",
				sourcePath,
				err,
			)
		}
		info, err := os.Stat(absoluteSourcePath)
		if errors.Is(err, os.ErrNotExist) {
			return domain.DockerImageSourceOverrides{}, fmt.Errorf(
				"invalid --image-source path %q: does not exist",
				absoluteSourcePath,
			)
		}
		if err != nil {
			return domain.DockerImageSourceOverrides{}, fmt.Errorf(
				"invalid --image-source path %q: %w",
				absoluteSourcePath,
				err,
			)
		}
		if !info.IsDir() {
			return domain.DockerImageSourceOverrides{}, fmt.Errorf(
				"invalid --image-source path %q: not a directory",
				absoluteSourcePath,
			)
		}
		sourcePathByImage[imageName] = absoluteSourcePath
	}
	return domain.NewDockerImageSourceOverrides(sourcePathByImage), nil
}

// ParseHelmChartOverrides converts repeated --helm-chart entries in
// "service:path" form into a HelmChartOverrides value. Duplicate service
// entries take the last value.
func ParseHelmChartOverrides(entries []string) (domain.HelmChartOverrides, error) {
	chartDirectoryByService := map[string]string{}
	for _, entry := range entries {
		serviceName, chartPath, found := strings.Cut(entry, ":")
		if !found {
			return domain.HelmChartOverrides{}, fmt.Errorf(
				"invalid --helm-chart value %q; expected format: service:path",
				entry,
			)
		}
		serviceName = strings.TrimSpace(serviceName)
		chartPath = strings.TrimSpace(chartPath)
		if serviceName == "" {
			return domain.HelmChartOverrides{}, fmt.Errorf(
				"invalid --helm-chart value %q; service name cannot be empty",
				entry,
			)
		}
		if chartPath == "" {
			return domain.HelmChartOverrides{}, fmt.Errorf(
				"invalid --helm-chart value %q; chart path cannot be empty",
				entry,
			)
		}
		if index := strings.IndexFunc(chartPath, func(r rune) bool {
			return r < 0x20 || r == 0x7f
		}); index >= 0 {
			return domain.HelmChartOverrides{}, fmt.Errorf(
				"invalid --helm-chart value %q; chart path contains a control character at byte %d",
				entry,
				index,
			)
		}
		expandedChartPath, err := expandLeadingTilde(chartPath)
		if err != nil {
			return domain.HelmChartOverrides{}, fmt.Errorf(
				"failed to resolve --helm-chart path %q: %w",
				chartPath,
				err,
			)
		}
		absoluteChartPath, err := filepath.Abs(expandedChartPath)
		if err != nil {
			return domain.HelmChartOverrides{}, fmt.Errorf(
				"failed to resolve --helm-chart path %q: %w",
				chartPath,
				err,
			)
		}
		info, err := os.Stat(absoluteChartPath)
		if errors.Is(err, os.ErrNotExist) {
			return domain.HelmChartOverrides{}, fmt.Errorf(
				"invalid --helm-chart path %q: does not exist",
				absoluteChartPath,
			)
		}
		if err != nil {
			return domain.HelmChartOverrides{}, fmt.Errorf(
				"invalid --helm-chart path %q: %w",
				absoluteChartPath,
				err,
			)
		}
		if !info.IsDir() {
			return domain.HelmChartOverrides{}, fmt.Errorf(
				"invalid --helm-chart path %q: not a directory",
				absoluteChartPath,
			)
		}
		chartDirectoryByService[serviceName] = absoluteChartPath
	}
	return domain.NewHelmChartOverrides(chartDirectoryByService), nil
}
