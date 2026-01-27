package handler

import (
	"errors"
	"testing"

	"dx/internal/core/domain"
	"dx/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSecretCommandHandler_HandleSet_Success(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	existingSecrets := []*domain.Secret{}

	terminalInput.On("IsTerminal").Return(true)
	terminalInput.On("ReadPassword", mock.Anything).Return("secret123", nil)
	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)
	secretsRepository.On("SaveSecrets", mock.MatchedBy(func(secrets []*domain.Secret) bool {
		// Verify the new secret was added with correct key and value
		if len(secrets) != 1 {
			return false
		}
		return secrets[0].Key == "DB_PASSWORD" && secrets[0].Value == "secret123"
	}), "test-context").Return(nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleSet("DB_PASSWORD")

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
	terminalInput.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleSet_UpdateExisting(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	existingSecrets := []*domain.Secret{
		{Key: "DB_PASSWORD", Value: "old-value"},
	}

	terminalInput.On("IsTerminal").Return(true)
	terminalInput.On("ReadPassword", mock.Anything).Return("new-value", nil)
	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)
	secretsRepository.On("SaveSecrets", mock.MatchedBy(func(secrets []*domain.Secret) bool {
		// Verify the secret was updated
		if len(secrets) != 1 {
			return false
		}
		return secrets[0].Key == "DB_PASSWORD" && secrets[0].Value == "new-value"
	}), "test-context").Return(nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleSet("DB_PASSWORD")

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
	terminalInput.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleSet_LoadContextNameError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	expectedErr := errors.New("load context name error")
	terminalInput.On("IsTerminal").Return(true)
	terminalInput.On("ReadPassword", mock.Anything).Return("secret123", nil)
	configRepository.On("LoadCurrentContextName").Return("", expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleSet("DB_PASSWORD")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	terminalInput.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleSet_LoadSecretsError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	expectedErr := errors.New("load secrets error")
	terminalInput.On("IsTerminal").Return(true)
	terminalInput.On("ReadPassword", mock.Anything).Return("secret123", nil)
	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(nil, expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleSet("DB_PASSWORD")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
	terminalInput.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleSet_SaveSecretsError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	existingSecrets := []*domain.Secret{}
	expectedErr := errors.New("save secrets error")

	terminalInput.On("IsTerminal").Return(true)
	terminalInput.On("ReadPassword", mock.Anything).Return("secret123", nil)
	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)
	secretsRepository.On("SaveSecrets", mock.Anything, "test-context").Return(expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleSet("DB_PASSWORD")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
	terminalInput.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleSet_NonTerminal(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	terminalInput.On("IsTerminal").Return(false)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleSet("DB_PASSWORD")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot read secret value: no terminal available")
	terminalInput.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleSet_ReadPasswordError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	terminalInput.On("IsTerminal").Return(true)
	terminalInput.On("ReadPassword", mock.Anything).Return("", errors.New("read error"))

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleSet("DB_PASSWORD")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read secret value")
	terminalInput.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleSet_EmptyValue(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	terminalInput.On("IsTerminal").Return(true)
	terminalInput.On("ReadPassword", mock.Anything).Return("", nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleSet("DB_PASSWORD")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "secret value cannot be empty")
	terminalInput.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleList_Success(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	configContext := &domain.ConfigurationContext{Name: "test-context"}
	secrets := []*domain.Secret{
		{Key: "DB_PASSWORD", Value: "secret123"},
		{Key: "API_KEY", Value: "api-key-value"},
	}

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(secrets, nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleList()

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleList_Empty(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	configContext := &domain.ConfigurationContext{Name: "test-context"}
	secrets := []*domain.Secret{}

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(secrets, nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleList()

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleList_LoadConfigError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	expectedErr := errors.New("load config error")
	configRepository.On("LoadCurrentConfigurationContext").Return(nil, expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleList()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleList_LoadSecretsError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	configContext := &domain.ConfigurationContext{Name: "test-context"}
	expectedErr := errors.New("load secrets error")

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(nil, expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleList()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleGet_Success(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	existingSecrets := []*domain.Secret{
		{Key: "DB_PASSWORD", Value: "secret123"},
		{Key: "API_KEY", Value: "api-key-value"},
	}

	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleGet("DB_PASSWORD")

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleGet_NotFound(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	existingSecrets := []*domain.Secret{
		{Key: "DB_PASSWORD", Value: "secret123"},
	}

	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleGet("NON_EXISTENT_KEY")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "secret 'NON_EXISTENT_KEY' not found")
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleGet_LoadContextNameError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	expectedErr := errors.New("load context name error")
	configRepository.On("LoadCurrentContextName").Return("", expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleGet("DB_PASSWORD")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleGet_LoadSecretsError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	expectedErr := errors.New("load secrets error")
	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(nil, expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleGet("DB_PASSWORD")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleDelete_Success(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	existingSecrets := []*domain.Secret{
		{Key: "DB_PASSWORD", Value: "secret123"},
		{Key: "API_KEY", Value: "api-key-value"},
	}

	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)
	secretsRepository.On("SaveSecrets", mock.MatchedBy(func(secrets []*domain.Secret) bool {
		// Verify DB_PASSWORD was deleted
		if len(secrets) != 1 {
			return false
		}
		return secrets[0].Key == "API_KEY"
	}), "test-context").Return(nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleDelete("DB_PASSWORD")

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleDelete_NonExistentKey(t *testing.T) {
	// Documents behavior: deleting a non-existent key silently succeeds
	// (the implementation filters and saves, even if the key wasn't present)
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	existingSecrets := []*domain.Secret{
		{Key: "DB_PASSWORD", Value: "secret123"},
	}

	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)
	secretsRepository.On("SaveSecrets", mock.MatchedBy(func(secrets []*domain.Secret) bool {
		// Verify existing secrets are preserved (non-existent key wasn't there to delete)
		if len(secrets) != 1 {
			return false
		}
		return secrets[0].Key == "DB_PASSWORD"
	}), "test-context").Return(nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleDelete("NON_EXISTENT_KEY")

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleDelete_LoadContextNameError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	expectedErr := errors.New("load context name error")
	configRepository.On("LoadCurrentContextName").Return("", expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleDelete("DB_PASSWORD")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleDelete_LoadSecretsError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	expectedErr := errors.New("load secrets error")
	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(nil, expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleDelete("DB_PASSWORD")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleDelete_SaveSecretsError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	existingSecrets := []*domain.Secret{
		{Key: "DB_PASSWORD", Value: "secret123"},
	}
	expectedErr := errors.New("save secrets error")

	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)
	secretsRepository.On("SaveSecrets", mock.Anything, "test-context").Return(expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleDelete("DB_PASSWORD")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

// HandleConfigure tests

func TestSecretCommandHandler_HandleConfigure_NoSecretsInConfig(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	configContext := &domain.ConfigurationContext{
		Name:    "test-context",
		Scripts: map[string]string{},
	}

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleConfigure(false)

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleConfigure_AllSecretsConfigured(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	configContext := &domain.ConfigurationContext{
		Name: "test-context",
		Scripts: map[string]string{
			"deploy": "echo {{.Secrets.DB_PASSWORD}}",
		},
	}
	existingSecrets := []*domain.Secret{
		{Key: "DB_PASSWORD", Value: "secret123"},
	}

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleConfigure(false)

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleConfigure_CheckOnly_MissingSecrets(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	configContext := &domain.ConfigurationContext{
		Name: "test-context",
		Scripts: map[string]string{
			"deploy": "{{.Secrets.DB_PASSWORD}} {{.Secrets.API_KEY}}",
		},
	}
	existingSecrets := []*domain.Secret{}

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleConfigure(true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "2 missing secrets")
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleConfigure_CheckOnly_PartialMissing(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	configContext := &domain.ConfigurationContext{
		Name: "test-context",
		Scripts: map[string]string{
			"deploy": "{{.Secrets.DB_PASSWORD}} {{.Secrets.API_KEY}}",
		},
	}
	existingSecrets := []*domain.Secret{
		{Key: "DB_PASSWORD", Value: "secret123"},
	}

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleConfigure(true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "1 missing secret")
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleConfigure_Interactive_NonTerminal(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	configContext := &domain.ConfigurationContext{
		Name: "test-context",
		Scripts: map[string]string{
			"deploy": "{{.Secrets.DB_PASSWORD}}",
		},
	}
	existingSecrets := []*domain.Secret{}

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)
	terminalInput.On("IsTerminal").Return(false)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleConfigure(false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interactive mode requires a terminal")
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
	terminalInput.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleConfigure_Interactive_Success(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	configContext := &domain.ConfigurationContext{
		Name: "test-context",
		Scripts: map[string]string{
			"deploy": "{{.Secrets.DB_PASSWORD}}",
		},
	}
	existingSecrets := []*domain.Secret{}

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)
	terminalInput.On("IsTerminal").Return(true)
	terminalInput.On("ReadPassword", mock.Anything).Return("secret123", nil)
	secretsRepository.On("SaveSecrets", mock.MatchedBy(func(secrets []*domain.Secret) bool {
		if len(secrets) != 1 {
			return false
		}
		return secrets[0].Key == "DB_PASSWORD" && secrets[0].Value == "secret123"
	}), "test-context").Return(nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleConfigure(false)

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
	terminalInput.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleConfigure_Interactive_SkipEmpty(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	configContext := &domain.ConfigurationContext{
		Name: "test-context",
		Scripts: map[string]string{
			"deploy": "{{.Secrets.DB_PASSWORD}}",
		},
	}
	existingSecrets := []*domain.Secret{}

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)
	terminalInput.On("IsTerminal").Return(true)
	terminalInput.On("ReadPassword", mock.Anything).Return("", nil) // Empty input
	// Note: SaveSecrets is NOT called when all secrets are skipped (no changes to save)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleConfigure(false)

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
	terminalInput.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleConfigure_Interactive_ReadPasswordError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	configContext := &domain.ConfigurationContext{
		Name: "test-context",
		Scripts: map[string]string{
			"deploy": "{{.Secrets.DB_PASSWORD}}",
		},
	}
	existingSecrets := []*domain.Secret{}

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)
	terminalInput.On("IsTerminal").Return(true)
	terminalInput.On("ReadPassword", mock.Anything).Return("", errors.New("read error"))

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleConfigure(false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read secret")
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
	terminalInput.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleConfigure_LoadConfigError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	expectedErr := errors.New("load config error")
	configRepository.On("LoadCurrentConfigurationContext").Return(nil, expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleConfigure(false)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleConfigure_LoadSecretsError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	configContext := &domain.ConfigurationContext{
		Name: "test-context",
		Scripts: map[string]string{
			"deploy": "{{.Secrets.DB_PASSWORD}}",
		},
	}
	expectedErr := errors.New("load secrets error")

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(nil, expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleConfigure(false)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleConfigure_FromHelmArgs(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	configContext := &domain.ConfigurationContext{
		Name: "test-context",
		Services: []domain.Service{
			{
				Name: "api",
				HelmArgs: []string{
					"--set=password={{.Secrets.DB_PASSWORD}}",
				},
			},
		},
	}
	existingSecrets := []*domain.Secret{}

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)
	terminalInput.On("IsTerminal").Return(true)
	terminalInput.On("ReadPassword", mock.Anything).Return("secret123", nil)
	secretsRepository.On("SaveSecrets", mock.MatchedBy(func(secrets []*domain.Secret) bool {
		if len(secrets) != 1 {
			return false
		}
		return secrets[0].Key == "DB_PASSWORD"
	}), "test-context").Return(nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleConfigure(false)

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
	terminalInput.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleConfigure_Interactive_SaveSecretsError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	configContext := &domain.ConfigurationContext{
		Name: "test-context",
		Scripts: map[string]string{
			"deploy": "{{.Secrets.DB_PASSWORD}}",
		},
	}
	existingSecrets := []*domain.Secret{}
	expectedErr := errors.New("save secrets error")

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)
	terminalInput.On("IsTerminal").Return(true)
	terminalInput.On("ReadPassword", mock.Anything).Return("secret123", nil)
	secretsRepository.On("SaveSecrets", mock.Anything, "test-context").Return(expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleConfigure(false)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
	terminalInput.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleConfigure_FromBuildArgs(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)
	terminalInput := new(testutil.MockTerminalInput)

	configContext := &domain.ConfigurationContext{
		Name: "test-context",
		Services: []domain.Service{
			{
				Name: "api",
				DockerImages: []domain.DockerImage{
					{
						Name: "api-image",
						BuildArgs: []string{
							"--build-arg=NPM_TOKEN={{.Secrets.NPM_TOKEN}}",
						},
					},
				},
			},
		},
	}
	existingSecrets := []*domain.Secret{}

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)
	terminalInput.On("IsTerminal").Return(true)
	terminalInput.On("ReadPassword", mock.Anything).Return("token123", nil)
	secretsRepository.On("SaveSecrets", mock.MatchedBy(func(secrets []*domain.Secret) bool {
		if len(secrets) != 1 {
			return false
		}
		return secrets[0].Key == "NPM_TOKEN"
	}), "test-context").Return(nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository, terminalInput)

	err := sut.HandleConfigure(false)

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
	terminalInput.AssertExpectations(t)
}
