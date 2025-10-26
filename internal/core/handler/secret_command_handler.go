package handler

import (
	"fmt"
	"sort"

	"dx/internal/core"
	"dx/internal/core/domain"
)

type SecretCommandHandler struct {
	secretsRepository core.SecretsRepository
	configRepository  core.ConfigRepository
}

func ProvideSecretCommandHandler(
	secretsRepository core.SecretsRepository,
	configRepository core.ConfigRepository,
) SecretCommandHandler {
	return SecretCommandHandler{
		secretsRepository: secretsRepository,
		configRepository:  configRepository,
	}
}

func (h *SecretCommandHandler) HandleSet(key string, value string) error {
	configContextName, err := h.configRepository.LoadCurrentContextName()
	secrets, err := h.secretsRepository.LoadSecrets(configContextName)
	if err != nil {
		return err
	}

	var secretExists = false
	for _, secret := range secrets {
		if secret.Key == key {
			secret.Value = value
			secretExists = true
		}
	}

	if !secretExists {
		secrets = append(
			secrets, &domain.Secret{
				Key:   key,
				Value: value,
			},
		)
	}

	return h.secretsRepository.SaveSecrets(secrets, configContextName)
}

func (h *SecretCommandHandler) HandleList() error {
	configContext, err := h.configRepository.LoadCurrentConfigurationContext()
	if err != nil {
		return err
	}
	secrets, err := h.secretsRepository.LoadSecrets(configContext.Name)
	if err != nil {
		return err
	}
	fmt.Println("Secrets:")

	// Sort secrets by key
	sort.Slice(
		secrets, func(i, j int) bool {
			return secrets[i].Key < secrets[j].Key
		},
	)
	for _, secret := range secrets {
		fmt.Printf("  %s: %s\n", secret.Key, secret.Value)
	}

	return nil
}

func (h *SecretCommandHandler) HandleDelete(key string) error {
	configContextName, err := h.configRepository.LoadCurrentContextName()
	secrets, err := h.secretsRepository.LoadSecrets(configContextName)
	if err != nil {
		return err
	}
	var newSecrets []*domain.Secret
	for _, secret := range secrets {
		if secret.Key != key {
			newSecrets = append(newSecrets, secret)
		}
	}
	return h.secretsRepository.SaveSecrets(newSecrets, configContextName)
}
