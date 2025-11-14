package handler

import (
	"testing"

	"dx/internal/core"
	"dx/internal/core/domain"
	"dx/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInstallCommandHandler_HandleInstallsAllServices(t *testing.T) {
	configContext := &domain.ConfigurationContext{
		Name: "Test",
		Services: []domain.Service{
			{
				Name:         "service-1",
				HelmRepoPath: "any-repo-1",
				HelmBranch:   "any-branch-1",
				Profiles:     []string{"all"},
			},
			{
				Name:         "service-2",
				HelmRepoPath: "any-repo-2",
				HelmBranch:   "any-branch-2",
				Profiles:     []string{"all"},
			},
		},
	}
	configRepository := new(testutil.MockConfigRepository)
	configRepository.On("LoadEnvKey", mock.Anything).Return("any-key", nil)
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	containerOrchestrator.On("CreateClusterEnvironmentKey").Return("any-key", nil)
	containerOrchestrator.On("InstallService", mock.Anything).Return(nil)
	containerOrchestrator.On("InstallDevProxy", mock.Anything).Return(nil)
	fileSystem := new(testutil.MockFileSystem)
	fileSystem.On("WriteFile", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	scm := new(testutil.MockScm)
	scm.On(
		"Download",
		configContext.Services[0].HelmRepoPath,
		configContext.Services[0].HelmBranch,
		configContext.Services[0].HelmPath,
	).Return(nil)
	scm.On(
		"Download",
		configContext.Services[1].HelmRepoPath,
		configContext.Services[1].HelmBranch,
		configContext.Services[1].HelmPath,
	).Return(nil)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerImageRepository.On("BuildImage", mock.Anything).Return(nil)
	devProxyManager := core.ProvideDevProxyManager(
		configRepository,
		fileSystem,
		containerImageRepository,
		containerOrchestrator,
	)
	environmentEnsurer := core.ProvideEnvironmentEnsurer(
		configRepository,
		containerOrchestrator,
	)
	sut := ProvideInstallCommandHandler(
		configRepository,
		containerImageRepository,
		containerOrchestrator,
		devProxyManager,
		environmentEnsurer,
		scm,
	)

	result := sut.Handle([]string{}, "all", false)

	assert.Nil(t, result)
	containerImageRepository.AssertExpectations(t)
	containerImageRepository.AssertNumberOfCalls(t, "BuildImage", 2)
	fileSystem.AssertExpectations(t)
	containerOrchestrator.AssertExpectations(t)
	containerOrchestrator.AssertNumberOfCalls(t, "InstallService", 2)
	containerOrchestrator.AssertNumberOfCalls(t, "InstallDevProxy", 1)
	scm.AssertNumberOfCalls(t, "Download", 2)
	scm.AssertExpectations(t)
}

func TestInstallCommandHandler_HandleInstallsOnlySelectedService(t *testing.T) {
	configContext := &domain.ConfigurationContext{
		Name: "Test",
		Services: []domain.Service{
			{
				Name:         "service-1",
				HelmRepoPath: "any-repo-1",
				HelmBranch:   "any-branch-1",
				Profiles:     []string{"default"},
			},
			{
				Name:         "service-2",
				HelmRepoPath: "any-repo-2",
				HelmBranch:   "any-branch-2",
				Profiles:     []string{"default"},
			},
		},
	}
	configRepository := new(testutil.MockConfigRepository)
	configRepository.On("LoadEnvKey", mock.Anything).Return("any-key", nil)
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	containerOrchestrator.On("CreateClusterEnvironmentKey").Return("any-key", nil)
	containerOrchestrator.On("InstallService", mock.Anything).Return(nil)
	containerOrchestrator.On("InstallDevProxy", mock.Anything).Return(nil)
	fileSystem := new(testutil.MockFileSystem)
	fileSystem.On("WriteFile", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	scm := new(testutil.MockScm)
	scm.On(
		"Download",
		configContext.Services[0].HelmRepoPath,
		configContext.Services[0].HelmBranch,
		configContext.Services[0].HelmPath,
	).Return(nil)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerImageRepository.On("BuildImage", mock.Anything).Return(nil)
	devProxyManager := core.ProvideDevProxyManager(
		configRepository,
		fileSystem,
		containerImageRepository,
		containerOrchestrator,
	)
	environmentEnsurer := core.ProvideEnvironmentEnsurer(
		configRepository,
		containerOrchestrator,
	)
	sut := ProvideInstallCommandHandler(
		configRepository,
		containerImageRepository,
		containerOrchestrator,
		devProxyManager,
		environmentEnsurer,
		scm,
	)

	result := sut.Handle([]string{"service-1"}, "all", false)

	assert.Nil(t, result)
	containerImageRepository.AssertExpectations(t)
	containerImageRepository.AssertNumberOfCalls(t, "BuildImage", 2)
	fileSystem.AssertExpectations(t)
	containerOrchestrator.AssertExpectations(t)
	containerOrchestrator.AssertNumberOfCalls(t, "InstallService", 1)
	containerOrchestrator.AssertNumberOfCalls(t, "InstallDevProxy", 1)
	scm.AssertNumberOfCalls(t, "Download", 1)
	scm.AssertExpectations(t)
}
