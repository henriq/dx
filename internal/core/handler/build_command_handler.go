package handler

import (
	"fmt"
	"slices"
	"strings"

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

		fmt.Println("\033[1;37mBuilding docker images\033[0m")
		fmt.Println()
		fmt.Printf(
			"\033[1;37m%-*s%-*s%-*s\033[0m\n",
			maxDockerImageNameLength,
			"Image",
			maxGitRepoPathLength,
			"Repo",
			maxGitBranchLength,
			"Ref",
		)

		for _, image := range dockerImagesToBuild {
			if image.GitRepoPath == "" && image.GitRef == "" {
				fmt.Println(image.Name)
			} else {
				fmt.Printf(
					"%-*s%-*s%-*s\n",
					maxDockerImageNameLength,
					image.Name,
					maxGitRepoPathLength,
					image.GitRepoPath,
					maxGitBranchLength,
					image.GitRef,
				)
			}

			if err := h.scm.Download(image.GitRepoPath, image.GitRef, image.Path); err != nil {
				return err
			}

			if err := h.containerImageRepository.BuildImage(image); err != nil {
				return err
			}
		}
		fmt.Println()
	}

	if len(dockerImagesToPull) > 0 {
		slices.Sort(dockerImagesToPull)

		fmt.Println("\033[1;37mPulling docker images\033[0m")
		fmt.Println()
		fmt.Printf(
			"\033[1;37m%s\033[0m\n",
			"Image",
		)

		for _, image := range dockerImagesToPull {
			fmt.Println(image)
			if err := h.containerImageRepository.PullImage(image); err != nil {
				return err
			}
		}
	}

	return nil
}
