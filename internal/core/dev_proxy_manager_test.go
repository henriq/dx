package core

import (
	"errors"
	"path/filepath"
	"testing"

	"dx/internal/core/domain"
	"dx/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createTestConfigContext() *domain.ConfigurationContext {
	return &domain.ConfigurationContext{
		Name: "test-context",
		LocalServices: []domain.LocalService{
			{
				Name:            "service-1",
				KubernetesPort:  8080,
				LocalPort:       3000,
				HealthCheckPath: "/health",
				Selector:        map[string]string{"app": "service-1"},
			},
		},
	}
}

func TestSaveConfiguration_Success(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	configContext := createTestConfigContext()
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/haproxy/haproxy.cfg", mock.Anything, mock.Anything).Return(nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/haproxy/Dockerfile", mock.Anything, mock.Anything).Return(nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/mitmproxy/Dockerfile", mock.Anything, mock.Anything).Return(nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/helm/Chart.yaml", mock.Anything, mock.Anything).Return(nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/helm/templates/dev-proxy.yaml", mock.Anything, mock.Anything).Return(nil)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.SaveConfiguration()

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	fileSystem.AssertExpectations(t)
	fileSystem.AssertNumberOfCalls(t, "WriteFile", 5)
}

func TestSaveConfiguration_LoadConfigError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	expectedErr := errors.New("load config error")
	configRepository.On("LoadCurrentConfigurationContext").Return(nil, expectedErr)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.SaveConfiguration()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
}

func TestSaveConfiguration_WriteHaproxyConfigError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	configContext := createTestConfigContext()
	expectedErr := errors.New("write haproxy config error")
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/haproxy/haproxy.cfg", mock.Anything, mock.Anything).Return(expectedErr)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.SaveConfiguration()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	fileSystem.AssertExpectations(t)
}

func TestSaveConfiguration_WriteHaproxyDockerfileError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	configContext := createTestConfigContext()
	expectedErr := errors.New("write haproxy dockerfile error")
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/haproxy/haproxy.cfg", mock.Anything, mock.Anything).Return(nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/haproxy/Dockerfile", mock.Anything, mock.Anything).Return(expectedErr)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.SaveConfiguration()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	fileSystem.AssertExpectations(t)
}

func TestSaveConfiguration_WriteMitmproxyDockerfileError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	configContext := createTestConfigContext()
	expectedErr := errors.New("write mitmproxy dockerfile error")
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/haproxy/haproxy.cfg", mock.Anything, mock.Anything).Return(nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/haproxy/Dockerfile", mock.Anything, mock.Anything).Return(nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/mitmproxy/Dockerfile", mock.Anything, mock.Anything).Return(expectedErr)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.SaveConfiguration()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	fileSystem.AssertExpectations(t)
}

func TestSaveConfiguration_WriteHelmChartError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	configContext := createTestConfigContext()
	expectedErr := errors.New("write helm chart error")
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/haproxy/haproxy.cfg", mock.Anything, mock.Anything).Return(nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/haproxy/Dockerfile", mock.Anything, mock.Anything).Return(nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/mitmproxy/Dockerfile", mock.Anything, mock.Anything).Return(nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/helm/Chart.yaml", mock.Anything, mock.Anything).Return(expectedErr)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.SaveConfiguration()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	fileSystem.AssertExpectations(t)
}

func TestSaveConfiguration_WriteDevProxyManifestsError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	configContext := createTestConfigContext()
	expectedErr := errors.New("write dev-proxy manifests error")
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/haproxy/haproxy.cfg", mock.Anything, mock.Anything).Return(nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/haproxy/Dockerfile", mock.Anything, mock.Anything).Return(nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/mitmproxy/Dockerfile", mock.Anything, mock.Anything).Return(nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/helm/Chart.yaml", mock.Anything, mock.Anything).Return(nil)
	fileSystem.On("WriteFile", "~/.dx/test-context/dev-proxy/helm/templates/dev-proxy.yaml", mock.Anything, mock.Anything).Return(expectedErr)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.SaveConfiguration()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	fileSystem.AssertExpectations(t)
}

func TestBuildDevProxy_Success(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	configContext := createTestConfigContext()
	homeDir := "/home/testuser"

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	fileSystem.On("HomeDir").Return(homeDir, nil)
	containerImageRepository.On("BuildImage", domain.DockerImage{
		Name:                     "henriq/haproxy-test-context",
		DockerfilePath:           "Dockerfile",
		BuildContextRelativePath: ".",
		Path:                     filepath.Join(homeDir, ".dx", "test-context", "dev-proxy", "haproxy"),
	}).Return(nil)
	containerImageRepository.On("BuildImage", domain.DockerImage{
		Name:                     "henriq/mitmproxy-test-context",
		DockerfilePath:           "Dockerfile",
		BuildContextRelativePath: ".",
		Path:                     filepath.Join(homeDir, ".dx", "test-context", "dev-proxy", "mitmproxy"),
	}).Return(nil)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.BuildDevProxy()

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	fileSystem.AssertExpectations(t)
	containerImageRepository.AssertExpectations(t)
	containerImageRepository.AssertNumberOfCalls(t, "BuildImage", 2)
}

func TestBuildDevProxy_LoadConfigError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	expectedErr := errors.New("load config error")
	configRepository.On("LoadCurrentConfigurationContext").Return(nil, expectedErr)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.BuildDevProxy()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
}

func TestBuildDevProxy_HomeDirError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	configContext := createTestConfigContext()
	expectedErr := errors.New("home dir error")

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	fileSystem.On("HomeDir").Return("", expectedErr)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.BuildDevProxy()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	fileSystem.AssertExpectations(t)
}

func TestBuildDevProxy_BuildFirstImageError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	configContext := createTestConfigContext()
	homeDir := "/home/testuser"
	buildErr := errors.New("build image error")

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	fileSystem.On("HomeDir").Return(homeDir, nil)
	containerImageRepository.On("BuildImage", domain.DockerImage{
		Name:                     "henriq/haproxy-test-context",
		DockerfilePath:           "Dockerfile",
		BuildContextRelativePath: ".",
		Path:                     filepath.Join(homeDir, ".dx", "test-context", "dev-proxy", "haproxy"),
	}).Return(buildErr)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.BuildDevProxy()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to build image")
	assert.Contains(t, err.Error(), "henriq/haproxy-test-context")
	configRepository.AssertExpectations(t)
	fileSystem.AssertExpectations(t)
	containerImageRepository.AssertExpectations(t)
	containerImageRepository.AssertNumberOfCalls(t, "BuildImage", 1)
}

func TestBuildDevProxy_BuildSecondImageError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	configContext := createTestConfigContext()
	homeDir := "/home/testuser"
	buildErr := errors.New("build image error")

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	fileSystem.On("HomeDir").Return(homeDir, nil)
	containerImageRepository.On("BuildImage", domain.DockerImage{
		Name:                     "henriq/haproxy-test-context",
		DockerfilePath:           "Dockerfile",
		BuildContextRelativePath: ".",
		Path:                     filepath.Join(homeDir, ".dx", "test-context", "dev-proxy", "haproxy"),
	}).Return(nil)
	containerImageRepository.On("BuildImage", domain.DockerImage{
		Name:                     "henriq/mitmproxy-test-context",
		DockerfilePath:           "Dockerfile",
		BuildContextRelativePath: ".",
		Path:                     filepath.Join(homeDir, ".dx", "test-context", "dev-proxy", "mitmproxy"),
	}).Return(buildErr)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.BuildDevProxy()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to build image")
	assert.Contains(t, err.Error(), "henriq/mitmproxy-test-context")
	configRepository.AssertExpectations(t)
	fileSystem.AssertExpectations(t)
	containerImageRepository.AssertExpectations(t)
	containerImageRepository.AssertNumberOfCalls(t, "BuildImage", 2)
}

func TestInstallDevProxy_Success(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	configContext := createTestConfigContext()
	homeDir := "/home/testuser"

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	fileSystem.On("HomeDir").Return(homeDir, nil)
	containerOrchestrator.On("InstallDevProxy", &domain.Service{
		Name:     "dev-proxy",
		HelmPath: filepath.Join(homeDir, ".dx", "test-context", "dev-proxy", "helm"),
	}).Return(nil)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.InstallDevProxy()

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	fileSystem.AssertExpectations(t)
	containerOrchestrator.AssertExpectations(t)
}

func TestInstallDevProxy_LoadConfigError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	expectedErr := errors.New("load config error")
	configRepository.On("LoadCurrentConfigurationContext").Return(nil, expectedErr)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.InstallDevProxy()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
}

func TestInstallDevProxy_HomeDirError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	configContext := createTestConfigContext()
	expectedErr := errors.New("home dir error")

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	fileSystem.On("HomeDir").Return("", expectedErr)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.InstallDevProxy()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	fileSystem.AssertExpectations(t)
}

func TestInstallDevProxy_InstallError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	configContext := createTestConfigContext()
	homeDir := "/home/testuser"
	expectedErr := errors.New("install error")

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	fileSystem.On("HomeDir").Return(homeDir, nil)
	containerOrchestrator.On("InstallDevProxy", &domain.Service{
		Name:     "dev-proxy",
		HelmPath: filepath.Join(homeDir, ".dx", "test-context", "dev-proxy", "helm"),
	}).Return(expectedErr)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.InstallDevProxy()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	fileSystem.AssertExpectations(t)
	containerOrchestrator.AssertExpectations(t)
}

func TestUninstallDevProxy_Success(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	configContext := createTestConfigContext()
	homeDir := "/home/testuser"

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	fileSystem.On("HomeDir").Return(homeDir, nil)
	containerOrchestrator.On("UninstallService", &domain.Service{
		Name:     "dev-proxy",
		HelmPath: filepath.Join(homeDir, ".dx", "test-context", "dev-proxy", "helm"),
	}).Return(nil)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.UninstallDevProxy()

	assert.NoError(t, err)
	configRepository.AssertExpectations(t)
	fileSystem.AssertExpectations(t)
	containerOrchestrator.AssertExpectations(t)
}

func TestUninstallDevProxy_LoadConfigError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	expectedErr := errors.New("load config error")
	configRepository.On("LoadCurrentConfigurationContext").Return(nil, expectedErr)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.UninstallDevProxy()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
}

func TestUninstallDevProxy_HomeDirError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	configContext := createTestConfigContext()
	expectedErr := errors.New("home dir error")

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	fileSystem.On("HomeDir").Return("", expectedErr)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.UninstallDevProxy()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	fileSystem.AssertExpectations(t)
}

func TestUninstallDevProxy_UninstallError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	fileSystem := new(testutil.MockFileSystem)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerOrchestrator := new(testutil.MockContainerOrchestrator)
	configGenerator := ProvideDevProxyConfigGenerator()

	configContext := createTestConfigContext()
	homeDir := "/home/testuser"
	expectedErr := errors.New("uninstall error")

	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	fileSystem.On("HomeDir").Return(homeDir, nil)
	containerOrchestrator.On("UninstallService", &domain.Service{
		Name:     "dev-proxy",
		HelmPath: filepath.Join(homeDir, ".dx", "test-context", "dev-proxy", "helm"),
	}).Return(expectedErr)

	sut := ProvideDevProxyManager(configRepository, fileSystem, containerImageRepository, containerOrchestrator, configGenerator)

	err := sut.UninstallDevProxy()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	configRepository.AssertExpectations(t)
	fileSystem.AssertExpectations(t)
	containerOrchestrator.AssertExpectations(t)
}
