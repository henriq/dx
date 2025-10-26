package handler

import (
	"fmt"
	"path/filepath"

	"dx/internal/core"
	"dx/internal/ports"
)

type GenEnvKeyCommandHandler struct {
	configRepository      core.ConfigRepository
	fileSystem            ports.FileSystem
	containerOrchestrator ports.ContainerOrchestrator
}

func ProvideGenEnvKeyCommandHandler(
	configRepository core.ConfigRepository,
	fileSystem ports.FileSystem,
	containerOrchestrator ports.ContainerOrchestrator,
) GenEnvKeyCommandHandler {
	return GenEnvKeyCommandHandler{
		configRepository:      configRepository,
		fileSystem:            fileSystem,
		containerOrchestrator: containerOrchestrator,
	}
}

func (h *GenEnvKeyCommandHandler) Handle() error {
	configContext, err := h.configRepository.LoadCurrentConfigurationContext()
	if err != nil {
		return err
	}
	envKey, err := h.containerOrchestrator.CreateClusterEnvironmentKey()
	if err != nil {
		return fmt.Errorf("failed to generate environment key: %v", err)
	}
	envKeyPath := filepath.Join("~", ".dx", configContext.Name, "env-key")
	return h.fileSystem.WriteFile(envKeyPath, []byte(envKey), ports.ReadWrite)
}
