package handler

import (
	"dx/internal/core"
	"dx/internal/ports"
	"encoding/json"
	"fmt"
)

type ContextCommandHandler struct {
	configRepository         core.ConfigRepository
	scm                      ports.Scm
	containerImageRepository ports.ContainerImageRepository
}

func ProvideContextCommandHandler(
	configRepository core.ConfigRepository,
	scm ports.Scm,
	containerImageRepository ports.ContainerImageRepository,
) ContextCommandHandler {
	return ContextCommandHandler{
		configRepository:         configRepository,
		scm:                      scm,
		containerImageRepository: containerImageRepository,
	}
}

func (h *ContextCommandHandler) HandleSet(contextName string) error {
	config, err := h.configRepository.LoadConfig()
	if err != nil {
		return err
	}
	if !config.ContextExists(contextName) {
		return fmt.Errorf("context not found: %s", contextName)
	}
	return h.configRepository.SaveCurrentContextName(contextName)
}

func (h *ContextCommandHandler) HandleList() error {
	config, err := h.configRepository.LoadConfig()
	if err != nil {
		return err
	}
	for _, context := range config.Contexts {
		fmt.Println(context.Name)
	}
	return nil
}

func (h *ContextCommandHandler) HandlePrint() error {
	configContext, err := h.configRepository.LoadCurrentConfigurationContext()
	if err != nil {
		return err
	}
	return prettyPrint(configContext)
}

func prettyPrint(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
