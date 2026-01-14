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

	existingSecrets := []*domain.Secret{}

	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)
	secretsRepository.On("SaveSecrets", mock.MatchedBy(func(secrets []*domain.Secret) bool {
		// Verify the new secret was added with correct key and value
		if len(secrets) != 1 {
			return false
		}
		return secrets[0].Key == "DB_PASSWORD" && secrets[0].Value == "secret123"
	}), "test-context").Return(nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository)

	err := sut.HandleSet("DB_PASSWORD", "secret123")

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleSet_UpdateExisting(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)

	existingSecrets := []*domain.Secret{
		{Key: "DB_PASSWORD", Value: "old-value"},
	}

	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)
	secretsRepository.On("SaveSecrets", mock.MatchedBy(func(secrets []*domain.Secret) bool {
		// Verify the secret was updated
		if len(secrets) != 1 {
			return false
		}
		return secrets[0].Key == "DB_PASSWORD" && secrets[0].Value == "new-value"
	}), "test-context").Return(nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository)

	err := sut.HandleSet("DB_PASSWORD", "new-value")

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleSet_LoadContextNameError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)

	expectedErr := errors.New("load context name error")
	configRepository.On("LoadCurrentContextName").Return("", expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository)

	err := sut.HandleSet("DB_PASSWORD", "secret123")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleSet_LoadSecretsError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)

	expectedErr := errors.New("load secrets error")
	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(nil, expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository)

	err := sut.HandleSet("DB_PASSWORD", "secret123")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleSet_SaveSecretsError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)

	existingSecrets := []*domain.Secret{}
	expectedErr := errors.New("save secrets error")

	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)
	secretsRepository.On("SaveSecrets", mock.Anything, "test-context").Return(expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository)

	err := sut.HandleSet("DB_PASSWORD", "secret123")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleList_Success(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)

	configContext := &domain.ConfigurationContext{Name: "test-context"}
	secrets := []*domain.Secret{
		{Key: "DB_PASSWORD", Value: "secret123"},
		{Key: "API_KEY", Value: "api-key-value"},
	}

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(secrets, nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository)

	err := sut.HandleList()

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleList_Empty(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)

	configContext := &domain.ConfigurationContext{Name: "test-context"}
	secrets := []*domain.Secret{}

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(secrets, nil)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository)

	err := sut.HandleList()

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleList_LoadConfigError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)

	expectedErr := errors.New("load config error")
	configRepository.On("LoadCurrentConfigurationContext").Return(nil, expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository)

	err := sut.HandleList()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleList_LoadSecretsError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)

	configContext := &domain.ConfigurationContext{Name: "test-context"}
	expectedErr := errors.New("load secrets error")

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(nil, expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository)

	err := sut.HandleList()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleDelete_Success(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)

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

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository)

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

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository)

	err := sut.HandleDelete("NON_EXISTENT_KEY")

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleDelete_LoadContextNameError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)

	expectedErr := errors.New("load context name error")
	configRepository.On("LoadCurrentContextName").Return("", expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository)

	err := sut.HandleDelete("DB_PASSWORD")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleDelete_LoadSecretsError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)

	expectedErr := errors.New("load secrets error")
	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(nil, expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository)

	err := sut.HandleDelete("DB_PASSWORD")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestSecretCommandHandler_HandleDelete_SaveSecretsError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)

	existingSecrets := []*domain.Secret{
		{Key: "DB_PASSWORD", Value: "secret123"},
	}
	expectedErr := errors.New("save secrets error")

	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(existingSecrets, nil)
	secretsRepository.On("SaveSecrets", mock.Anything, "test-context").Return(expectedErr)

	sut := ProvideSecretCommandHandler(secretsRepository, configRepository)

	err := sut.HandleDelete("DB_PASSWORD")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}
