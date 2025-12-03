package container_image_repository

import (
	"fmt"

	"dx/internal/core"
	"dx/internal/core/domain"
	"dx/internal/ports"

	"os/exec"
	"path/filepath"
	"strings"
)

type DockerRepository struct {
	configRepository  core.ConfigRepository
	secretsRepository core.SecretsRepository
	templater         ports.Templater
}

func ProvideDockerRepository(
	configRepository core.ConfigRepository,
	secretsRepository core.SecretsRepository,
	templater ports.Templater,
) *DockerRepository {
	return &DockerRepository{
		configRepository:  configRepository,
		secretsRepository: secretsRepository,
		templater:         templater,
	}
}

func (d *DockerRepository) BuildImage(image domain.DockerImage) error {
	contextPath := filepath.Join(image.Path, image.BuildContextRelativePath)

	// Determine dockerfile path: use stdin ("-") for override, or file path
	var dockerfilePath string
	var dockerfileContent string
	if image.DockerfileOverride != "" {
		dockerfilePath = "-"
		dockerfileContent = image.DockerfileOverride
	} else {
		dockerfilePath = filepath.Join(image.Path, image.DockerfilePath)
	}

	// Execute docker build command
	args := []string{"build", "-t", image.Name, "-f", dockerfilePath}

	templateValues, err := core.CreateTemplatingValues(d.configRepository, d.secretsRepository)
	if err != nil {
		return err
	}

	for i, arg := range image.BuildArgs {
		renderedArg, err := d.templater.Render(arg, fmt.Sprintf("build-args.%d", i), templateValues)
		if err != nil {
			return err
		}
		args = append(args, renderedArg)
	}

	// Add context path as the last argument
	args = append(args, contextPath)

	output := strings.Builder{}

	cmd := exec.Command("docker", args...)
	cmd.Stdout = &output
	cmd.Stderr = &output

	// If using dockerfile override, pipe the content via stdin
	if dockerfileContent != "" {
		cmd.Stdin = strings.NewReader(dockerfileContent)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build image: %v\n%s", err, output.String())
	}

	return nil
}

// PullImage pulls a Docker image from a registry
func (d *DockerRepository) PullImage(imageName string) error {
	output := strings.Builder{}
	cmd := exec.Command("docker", "pull", imageName)
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull image: %v\n%s", err, output.String())
	}

	return nil
}
