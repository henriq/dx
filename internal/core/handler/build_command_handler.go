package handler

import (
	"fmt"
	"slices"
	"strings"

	"dx/internal/cli/output"
	"dx/internal/cli/progress"
	"dx/internal/core"
	"dx/internal/core/domain"
	"dx/internal/ports"
)

type BuildCommandHandler struct {
	configRepository         core.ConfigRepository
	scm                      ports.Scm
	containerImageRepository ports.ContainerImageRepository
}

func ProvideBuildCommandHandler(
	configRepository core.ConfigRepository,
	scm ports.Scm,
	containerImageRepository ports.ContainerImageRepository,
) BuildCommandHandler {
	return BuildCommandHandler{
		configRepository:         configRepository,
		scm:                      scm,
		containerImageRepository: containerImageRepository,
	}
}

func (h *BuildCommandHandler) Handle(services []string, selectedProfile string) error {
	maxDockerImageNameLength := 0
	maxGitRepoPathLength := 0
	maxGitBranchLength := 0
	var dockerImagesToBuild []domain.DockerImage
	var dockerImagesToPull []string

	configContext, err := h.configRepository.LoadCurrentConfigurationContext()
	if err != nil {
		return err
	}

	for _, service := range configContext.Services {
		if len(services) == 0 && !slices.Contains(service.Profiles, selectedProfile) {
			continue
		}

		if len(services) > 0 && !slices.ContainsFunc(services, func(s string) bool { return s == service.Name }) {
			continue
		}

		for _, image := range service.DockerImages {
			if len(image.Name)+2 > maxDockerImageNameLength {
				maxDockerImageNameLength = len(image.Name) + 2
			}
			if len(image.GitRepoPath)+2 > maxGitRepoPathLength {
				maxGitRepoPathLength = len(image.GitRepoPath) + 2
			}
			if len(image.GitRef) > maxGitBranchLength {
				maxGitBranchLength = len(image.GitRef)
			}
			dockerImagesToBuild = append(dockerImagesToBuild, image)
		}

		dockerImagesToPull = append(dockerImagesToPull, service.RemoteImages...)
	}

	if len(dockerImagesToBuild) > 0 {
		slices.SortFunc(
			dockerImagesToBuild, func(a, b domain.DockerImage) int {
				return strings.Compare(a.Name, b.Name)
			},
		)

		output.PrintHeader("Building docker images")
		fmt.Println()

		// Create progress tracker with repo/ref info
		imageNames := make([]string, len(dockerImagesToBuild))
		imageInfos := make([]string, len(dockerImagesToBuild))
		for i, img := range dockerImagesToBuild {
			imageNames[i] = img.Name
			if img.GitRepoPath != "" && img.GitRef != "" {
				imageInfos[i] = fmt.Sprintf("%s @ %s", img.GitRepoPath, img.GitRef)
			}
		}
		tracker := progress.NewTrackerWithInfo(imageNames, imageInfos)
		tracker.Start()

		var buildErr error
		for i, image := range dockerImagesToBuild {
			tracker.StartItem(i)

			// Show dockerfile override note (only in non-TTY mode, TTY shows spinner)
			if image.DockerfileOverride != "" {
				output.PrintSecondary("Using inline Dockerfile from configuration")
			}

			// Download source
			if err := h.scm.Download(image.GitRepoPath, image.GitRef, image.Path); err != nil {
				tracker.CompleteItem(i, err)
				tracker.PrintItemComplete(i)
				buildErr = err
				break
			}

			// Build image
			if err := h.containerImageRepository.BuildImage(image); err != nil {
				tracker.CompleteItem(i, err)
				tracker.PrintItemComplete(i)
				buildErr = err
				break
			}

			tracker.CompleteItem(i, nil)
			tracker.PrintItemComplete(i)
		}

		tracker.Stop()

		if buildErr != nil {
			return buildErr
		}

		fmt.Println()
		output.PrintSuccess(fmt.Sprintf("Built %d images", len(dockerImagesToBuild)))
		fmt.Println()
	}

	if len(dockerImagesToPull) > 0 {
		slices.Sort(dockerImagesToPull)

		output.PrintHeader("Pulling docker images")
		fmt.Println()

		// Create progress tracker for pulls
		tracker := progress.NewTracker(dockerImagesToPull)
		tracker.Start()

		var pullErr error
		for i, image := range dockerImagesToPull {
			tracker.StartItem(i)

			if err := h.containerImageRepository.PullImage(image); err != nil {
				tracker.CompleteItem(i, err)
				tracker.PrintItemComplete(i)
				pullErr = err
				break
			}

			tracker.CompleteItem(i, nil)
			tracker.PrintItemComplete(i)
		}

		tracker.Stop()

		if pullErr != nil {
			return pullErr
		}

		fmt.Println()
		output.PrintSuccess(fmt.Sprintf("Pulled %d images", len(dockerImagesToPull)))
	}

	return nil
}
