package handler

import (
	"errors"
	"testing"

	"dx/internal/core/domain"
	"dx/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func TestShowVarsCommandHandler_Handle_Success(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)

	configContext := &domain.ConfigurationContext{
		Name: "test-context",
		Services: []domain.Service{
			{Name: "service-1", Path: "/path/to/service"},
		},
	}
	secrets := []*domain.Secret{
		{Key: "DB_PASSWORD", Value: "secret123"},
	}

	// For CreateTemplatingValues
	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(secrets, nil)

	sut := ProvideShowVarsCommandHandler(secretsRepository, configRepository)

	err := sut.Handle()

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestShowVarsCommandHandler_Handle_NoSecrets(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)

	configContext := &domain.ConfigurationContext{
		Name:     "test-context",
		Services: []domain.Service{},
	}
	secrets := []*domain.Secret{}

	// For CreateTemplatingValues
	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(secrets, nil)

	sut := ProvideShowVarsCommandHandler(secretsRepository, configRepository)

	err := sut.Handle()

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}

func TestShowVarsCommandHandler_Handle_LoadContextNameError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)

	expectedErr := errors.New("load context name error")
	configRepository.On("LoadCurrentContextName").Return("", expectedErr)

	sut := ProvideShowVarsCommandHandler(secretsRepository, configRepository)

	err := sut.Handle()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
}

func TestShowVarsCommandHandler_Handle_LoadConfigError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)

	expectedErr := errors.New("load config error")
	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	configRepository.On("LoadCurrentConfigurationContext").Return(nil, expectedErr)

	sut := ProvideShowVarsCommandHandler(secretsRepository, configRepository)

	err := sut.Handle()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
}

func TestShowVarsCommandHandler_Handle_LoadSecretsError(t *testing.T) {
	secretsRepository := new(testutil.MockSecretsRepository)
	configRepository := new(testutil.MockConfigRepository)

	configContext := &domain.ConfigurationContext{Name: "test-context"}
	expectedErr := errors.New("load secrets error")

	configRepository.On("LoadCurrentContextName").Return("test-context", nil)
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	secretsRepository.On("LoadSecrets", "test-context").Return(nil, expectedErr)

	sut := ProvideShowVarsCommandHandler(secretsRepository, configRepository)

	err := sut.Handle()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	secretsRepository.AssertExpectations(t)
}
