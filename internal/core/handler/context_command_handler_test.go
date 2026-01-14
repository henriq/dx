package handler

import (
	"errors"
	"testing"

	"dx/internal/core/domain"
	"dx/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func TestContextCommandHandler_HandleSet_Success(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	scm := new(testutil.MockScm)
	containerImageRepository := new(testutil.MockContainerImageRepository)

	config := &domain.Config{
		Contexts: []domain.ConfigurationContext{
			{Name: "default"},
			{Name: "production"},
		},
	}

	configRepository.On("LoadConfig").Return(config, nil)
	configRepository.On("SaveCurrentContextName", "production").Return(nil)

	sut := ProvideContextCommandHandler(configRepository, scm, containerImageRepository)

	err := sut.HandleSet("production")

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
}

func TestContextCommandHandler_HandleSet_LoadConfigError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	scm := new(testutil.MockScm)
	containerImageRepository := new(testutil.MockContainerImageRepository)

	expectedErr := errors.New("load config error")
	configRepository.On("LoadConfig").Return(nil, expectedErr)

	sut := ProvideContextCommandHandler(configRepository, scm, containerImageRepository)

	err := sut.HandleSet("production")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
}

func TestContextCommandHandler_HandleSet_ContextNotFound(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	scm := new(testutil.MockScm)
	containerImageRepository := new(testutil.MockContainerImageRepository)

	config := &domain.Config{
		Contexts: []domain.ConfigurationContext{
			{Name: "default"},
		},
	}

	configRepository.On("LoadConfig").Return(config, nil)

	sut := ProvideContextCommandHandler(configRepository, scm, containerImageRepository)

	err := sut.HandleSet("non-existent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context not found: non-existent")
	configRepository.AssertExpectations(t)
}

func TestContextCommandHandler_HandleSet_SaveError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	scm := new(testutil.MockScm)
	containerImageRepository := new(testutil.MockContainerImageRepository)

	config := &domain.Config{
		Contexts: []domain.ConfigurationContext{
			{Name: "default"},
			{Name: "production"},
		},
	}
	expectedErr := errors.New("save error")

	configRepository.On("LoadConfig").Return(config, nil)
	configRepository.On("SaveCurrentContextName", "production").Return(expectedErr)

	sut := ProvideContextCommandHandler(configRepository, scm, containerImageRepository)

	err := sut.HandleSet("production")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
}

func TestContextCommandHandler_HandleList_Success(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	scm := new(testutil.MockScm)
	containerImageRepository := new(testutil.MockContainerImageRepository)

	config := &domain.Config{
		Contexts: []domain.ConfigurationContext{
			{Name: "default"},
			{Name: "production"},
		},
	}

	configRepository.On("LoadConfig").Return(config, nil)
	configRepository.On("LoadCurrentContextName").Return("default", nil)

	sut := ProvideContextCommandHandler(configRepository, scm, containerImageRepository)

	err := sut.HandleList()

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
}

func TestContextCommandHandler_HandleList_Empty(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	scm := new(testutil.MockScm)
	containerImageRepository := new(testutil.MockContainerImageRepository)

	config := &domain.Config{
		Contexts: []domain.ConfigurationContext{},
	}

	configRepository.On("LoadConfig").Return(config, nil)
	configRepository.On("LoadCurrentContextName").Return("", nil)

	sut := ProvideContextCommandHandler(configRepository, scm, containerImageRepository)

	err := sut.HandleList()

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
}

func TestContextCommandHandler_HandleList_LoadConfigError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	scm := new(testutil.MockScm)
	containerImageRepository := new(testutil.MockContainerImageRepository)

	expectedErr := errors.New("load config error")
	configRepository.On("LoadConfig").Return(nil, expectedErr)

	sut := ProvideContextCommandHandler(configRepository, scm, containerImageRepository)

	err := sut.HandleList()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
}

func TestContextCommandHandler_HandleList_LoadCurrentContextNameError(t *testing.T) {
	// Documents behavior: LoadCurrentContextName error is intentionally ignored
	// The list still displays but without marking any context as current
	configRepository := new(testutil.MockConfigRepository)
	scm := new(testutil.MockScm)
	containerImageRepository := new(testutil.MockContainerImageRepository)

	config := &domain.Config{
		Contexts: []domain.ConfigurationContext{
			{Name: "default"},
			{Name: "production"},
		},
	}

	configRepository.On("LoadConfig").Return(config, nil)
	configRepository.On("LoadCurrentContextName").Return("", errors.New("context name error"))

	sut := ProvideContextCommandHandler(configRepository, scm, containerImageRepository)

	err := sut.HandleList()

	// Should succeed even when LoadCurrentContextName fails
	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
}

func TestContextCommandHandler_HandlePrint_Success(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	scm := new(testutil.MockScm)
	containerImageRepository := new(testutil.MockContainerImageRepository)

	configContext := &domain.ConfigurationContext{
		Name: "test-context",
		Services: []domain.Service{
			{Name: "service-1"},
		},
	}

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)

	sut := ProvideContextCommandHandler(configRepository, scm, containerImageRepository)

	err := sut.HandlePrint()

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
}

func TestContextCommandHandler_HandlePrint_LoadConfigError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	scm := new(testutil.MockScm)
	containerImageRepository := new(testutil.MockContainerImageRepository)

	expectedErr := errors.New("load config error")
	configRepository.On("LoadCurrentConfigurationContext").Return(nil, expectedErr)

	sut := ProvideContextCommandHandler(configRepository, scm, containerImageRepository)

	err := sut.HandlePrint()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
}
