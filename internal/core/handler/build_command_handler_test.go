package handler

import (
	"testing"

	"pilot/internal/core/domain"
	"pilot/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBuildCommandHandler_HandleBuildsAllServices(t *testing.T) {
	configContext := &domain.ConfigurationContext{
		Services: []domain.Service{
			{
				Name: "service-1",
				DockerImages: []domain.DockerImage{
					{
						Name:                     "any-image",
						DockerfilePath:           ".",
						BuildContextRelativePath: "",
						GitRepoPath:              "any-repo",
						GitRef:                   "any-branch",
					},
				},
				Profiles:     []string{"all"},
				RemoteImages: []string{"any-image"},
			},
			{
				Name: "service-2",
				DockerImages: []domain.DockerImage{
					{
						Name:                     "any-image-2",
						DockerfilePath:           ".",
						BuildContextRelativePath: "",
						GitRepoPath:              "any-repo-2",
						GitRef:                   "any-branch-2",
					},
				},
				Profiles:     []string{"all"},
				RemoteImages: []string{"any-image-2"},
			},
		},
	}
	configRepository := new(testutil.MockConfigRepository)
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	scm := new(testutil.MockScm)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	scm.On(
		"Download",
		configContext.Services[0].DockerImages[0].GitRepoPath,
		configContext.Services[0].DockerImages[0].GitRef,
		configContext.Services[0].DockerImages[0].Path,
	).Return(nil)
	containerImageRepository.On("PullImage", configContext.Services[0].RemoteImages[0]).Return(nil)
	containerImageRepository.On("BuildImage", configContext.Services[0].DockerImages[0]).Return(nil)
	scm.On(
		"Download",
		configContext.Services[1].DockerImages[0].GitRepoPath,
		configContext.Services[1].DockerImages[0].GitRef,
		configContext.Services[1].DockerImages[0].Path,
	).Return(nil)
	containerImageRepository.On("PullImage", configContext.Services[1].RemoteImages[0]).Return(nil)
	containerImageRepository.On("BuildImage", configContext.Services[1].DockerImages[0]).Return(nil)

	sut := BuildCommandHandler{
		configRepository:         configRepository,
		scm:                      scm,
		containerImageRepository: containerImageRepository,
	}

	result := sut.Handle([]string{}, "all", domain.DockerImageSourceOverrides{})

	assert.Nil(t, result)
	scm.AssertExpectations(t)
	containerImageRepository.AssertExpectations(t)
}

func TestBuildCommandHandler_HandleBuildsOnlySelectedService(t *testing.T) {
	configContext := &domain.ConfigurationContext{
		Services: []domain.Service{
			{
				Name: "service-1",
				DockerImages: []domain.DockerImage{
					{
						Name:                     "any-image",
						DockerfilePath:           ".",
						BuildContextRelativePath: "",
						GitRepoPath:              "any-repo",
						GitRef:                   "any-branch",
					},
				},
				Profiles:     []string{"all"},
				RemoteImages: []string{"any-image"},
			},
			{
				Name: "service-2",
				DockerImages: []domain.DockerImage{
					{
						Name:                     "any-image-2",
						DockerfilePath:           ".",
						BuildContextRelativePath: "",
						GitRepoPath:              "any-repo-2",
						GitRef:                   "any-branch-2",
					},
				},
				Profiles:     []string{"all"},
				RemoteImages: []string{"any-image-2"},
			},
		},
	}
	configRepository := new(testutil.MockConfigRepository)
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	scm := new(testutil.MockScm)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	scm.On(
		"Download",
		configContext.Services[0].DockerImages[0].GitRepoPath,
		configContext.Services[0].DockerImages[0].GitRef,
		configContext.Services[0].DockerImages[0].Path,
	).Return(nil)
	containerImageRepository.On("PullImage", configContext.Services[0].RemoteImages[0]).Return(nil)
	containerImageRepository.On("BuildImage", configContext.Services[0].DockerImages[0]).Return(nil)

	sut := NewBuildCommandHandler(
		configRepository,
		scm,
		containerImageRepository,
	)

	result := sut.Handle([]string{configContext.Services[0].Name}, "default", domain.DockerImageSourceOverrides{})

	assert.Nil(t, result)
	scm.AssertExpectations(t)
	containerImageRepository.AssertExpectations(t)
}

func TestBuildCommandHandler_HandleBuildsOnlyServicesInSelectedProfile(t *testing.T) {
	configContext := &domain.ConfigurationContext{
		Services: []domain.Service{
			{
				Name: "service-1",
				DockerImages: []domain.DockerImage{
					{
						Name:                     "any-image",
						DockerfilePath:           ".",
						BuildContextRelativePath: "",
						GitRepoPath:              "any-repo",
						GitRef:                   "any-branch",
					},
				},
				Profiles:     []string{"selected"},
				RemoteImages: []string{"any-image"},
			},
			{
				Name: "service-2",
				DockerImages: []domain.DockerImage{
					{
						Name:                     "any-image-2",
						DockerfilePath:           ".",
						BuildContextRelativePath: "",
						GitRepoPath:              "any-repo-2",
						GitRef:                   "any-branch-2",
					},
				},
				Profiles:     []string{"not-selected"},
				RemoteImages: []string{"any-image-2"},
			},
		},
	}
	configRepository := new(testutil.MockConfigRepository)
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	scm := new(testutil.MockScm)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	scm.On(
		"Download",
		configContext.Services[0].DockerImages[0].GitRepoPath,
		configContext.Services[0].DockerImages[0].GitRef,
		configContext.Services[0].DockerImages[0].Path,
	).Return(nil)
	containerImageRepository.On("PullImage", configContext.Services[0].RemoteImages[0]).Return(nil)
	containerImageRepository.On("BuildImage", configContext.Services[0].DockerImages[0]).Return(nil)

	sut := NewBuildCommandHandler(
		configRepository,
		scm,
		containerImageRepository,
	)

	result := sut.Handle([]string{}, "selected", domain.DockerImageSourceOverrides{})

	assert.Nil(t, result)
	scm.AssertExpectations(t)
	containerImageRepository.AssertExpectations(t)
}

func TestBuildCommandHandler_Handle_LoadConfigError(t *testing.T) {
	configRepository := new(testutil.MockConfigRepository)
	configRepository.On("LoadCurrentConfigurationContext").Return(nil, assert.AnError)
	scm := new(testutil.MockScm)
	containerImageRepository := new(testutil.MockContainerImageRepository)

	sut := NewBuildCommandHandler(
		configRepository,
		scm,
		containerImageRepository,
	)

	result := sut.Handle([]string{}, "all", domain.DockerImageSourceOverrides{})

	assert.ErrorIs(t, result, assert.AnError)
	scm.AssertNotCalled(t, "Download", mock.Anything, mock.Anything, mock.Anything)
	containerImageRepository.AssertNotCalled(t, "BuildImage", mock.Anything)
}

func TestBuildCommandHandler_Handle_DownloadError(t *testing.T) {
	configContext := &domain.ConfigurationContext{
		Services: []domain.Service{
			{
				Name: "service-1",
				DockerImages: []domain.DockerImage{
					{
						Name:                     "any-image",
						DockerfilePath:           ".",
						BuildContextRelativePath: "",
						GitRepoPath:              "any-repo",
						GitRef:                   "any-branch",
					},
				},
				Profiles: []string{"all"},
			},
		},
	}
	configRepository := new(testutil.MockConfigRepository)
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	scm := new(testutil.MockScm)
	scm.On(
		"Download",
		configContext.Services[0].DockerImages[0].GitRepoPath,
		configContext.Services[0].DockerImages[0].GitRef,
		configContext.Services[0].DockerImages[0].Path,
	).Return(assert.AnError)
	containerImageRepository := new(testutil.MockContainerImageRepository)

	sut := NewBuildCommandHandler(
		configRepository,
		scm,
		containerImageRepository,
	)

	result := sut.Handle([]string{}, "all", domain.DockerImageSourceOverrides{})

	assert.ErrorIs(t, result, assert.AnError)
	containerImageRepository.AssertNotCalled(t, "BuildImage", mock.Anything)
}

func TestBuildCommandHandler_Handle_BuildImageError(t *testing.T) {
	configContext := &domain.ConfigurationContext{
		Services: []domain.Service{
			{
				Name: "service-1",
				DockerImages: []domain.DockerImage{
					{
						Name:                     "any-image",
						DockerfilePath:           ".",
						BuildContextRelativePath: "",
						GitRepoPath:              "any-repo",
						GitRef:                   "any-branch",
					},
				},
				Profiles: []string{"all"},
			},
		},
	}
	configRepository := new(testutil.MockConfigRepository)
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	scm := new(testutil.MockScm)
	scm.On(
		"Download",
		configContext.Services[0].DockerImages[0].GitRepoPath,
		configContext.Services[0].DockerImages[0].GitRef,
		configContext.Services[0].DockerImages[0].Path,
	).Return(nil)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerImageRepository.On("BuildImage", configContext.Services[0].DockerImages[0]).Return(assert.AnError)

	sut := NewBuildCommandHandler(
		configRepository,
		scm,
		containerImageRepository,
	)

	result := sut.Handle([]string{}, "all", domain.DockerImageSourceOverrides{})

	assert.ErrorIs(t, result, assert.AnError)
	scm.AssertExpectations(t)
	containerImageRepository.AssertExpectations(t)
}

func TestBuildCommandHandler_Handle_SkipsDownloadWhenImageSourceOverridden(t *testing.T) {
	configContext := &domain.ConfigurationContext{
		Services: []domain.Service{
			{
				Name: "service-1",
				DockerImages: []domain.DockerImage{
					{
						Name:                     "any-image",
						DockerfilePath:           "Dockerfile",
						BuildContextRelativePath: "sub",
						GitRepoPath:              "any-repo",
						GitRef:                   "any-branch",
						Path:                     "/cache/any-image",
					},
				},
				Profiles: []string{"all"},
			},
		},
	}
	configRepository := new(testutil.MockConfigRepository)
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	scm := new(testutil.MockScm)
	containerImageRepository := new(testutil.MockContainerImageRepository)

	overrides := domain.NewDockerImageSourceOverrides(map[string]string{"any-image": "/local/checkout"})

	expectedBuiltImage := configContext.Services[0].DockerImages[0]
	expectedBuiltImage.Path = "/local/checkout"
	containerImageRepository.On("BuildImage", expectedBuiltImage).Return(nil)

	sut := NewBuildCommandHandler(configRepository, scm, containerImageRepository)

	result := sut.Handle([]string{}, "all", overrides)

	assert.Nil(t, result)
	scm.AssertNotCalled(t, "Download", mock.Anything, mock.Anything, mock.Anything)
	containerImageRepository.AssertExpectations(t)
}

func TestBuildCommandHandler_Handle_RejectsUnknownImageSourceOverride(t *testing.T) {
	configContext := &domain.ConfigurationContext{
		Services: []domain.Service{
			{
				Name: "service-1",
				DockerImages: []domain.DockerImage{
					{Name: "any-image", DockerfilePath: ".", BuildContextRelativePath: "", GitRepoPath: "r", GitRef: "b"},
				},
				Profiles: []string{"all"},
			},
		},
	}
	configRepository := new(testutil.MockConfigRepository)
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	scm := new(testutil.MockScm)
	containerImageRepository := new(testutil.MockContainerImageRepository)

	overrides := domain.NewDockerImageSourceOverrides(map[string]string{"unknown-image": "/local"})

	sut := NewBuildCommandHandler(configRepository, scm, containerImageRepository)

	result := sut.Handle([]string{}, "all", overrides)

	assert.Error(t, result)
	assert.Contains(t, result.Error(), "unknown-image")
	scm.AssertNotCalled(t, "Download", mock.Anything, mock.Anything, mock.Anything)
	containerImageRepository.AssertNotCalled(t, "BuildImage", mock.Anything)
}

func TestBuildCommandHandler_Handle_OverridesOneImageInMultiImageService(t *testing.T) {
	configContext := &domain.ConfigurationContext{
		Services: []domain.Service{
			{
				Name: "service-1",
				DockerImages: []domain.DockerImage{
					{Name: "overridden-image", DockerfilePath: ".", BuildContextRelativePath: "", GitRepoPath: "r1", GitRef: "b1", Path: "/cache/overridden"},
					{Name: "untouched-image", DockerfilePath: ".", BuildContextRelativePath: "", GitRepoPath: "r2", GitRef: "b2", Path: "/cache/untouched"},
				},
				Profiles: []string{"all"},
			},
		},
	}
	configRepository := new(testutil.MockConfigRepository)
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	scm := new(testutil.MockScm)
	scm.On("Download", "r2", "b2", "/cache/untouched").Return(nil)
	containerImageRepository := new(testutil.MockContainerImageRepository)

	overrides := domain.NewDockerImageSourceOverrides(map[string]string{"overridden-image": "/local"})

	expectedOverridden := configContext.Services[0].DockerImages[0]
	expectedOverridden.Path = "/local"
	containerImageRepository.On("BuildImage", expectedOverridden).Return(nil)
	containerImageRepository.On("BuildImage", configContext.Services[0].DockerImages[1]).Return(nil)

	sut := NewBuildCommandHandler(configRepository, scm, containerImageRepository)

	result := sut.Handle([]string{}, "all", overrides)

	assert.Nil(t, result)
	scm.AssertNumberOfCalls(t, "Download", 1)
	scm.AssertNotCalled(t, "Download", "r1", "b1", mock.Anything)
	containerImageRepository.AssertExpectations(t)
}

func TestBuildCommandHandler_Handle_WarnsWhenImageSourceOverrideOutOfScope(t *testing.T) {
	configContext := &domain.ConfigurationContext{
		Services: []domain.Service{
			{
				Name: "in-scope",
				DockerImages: []domain.DockerImage{
					{Name: "in-scope-image", DockerfilePath: ".", BuildContextRelativePath: "", GitRepoPath: "r1", GitRef: "b1"},
				},
				Profiles: []string{"selected"},
			},
			{
				Name: "out-of-scope",
				DockerImages: []domain.DockerImage{
					{Name: "out-of-scope-image", DockerfilePath: ".", BuildContextRelativePath: "", GitRepoPath: "r2", GitRef: "b2"},
				},
				Profiles: []string{"not-selected"},
			},
		},
	}
	configRepository := new(testutil.MockConfigRepository)
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	scm := new(testutil.MockScm)
	scm.On("Download", "r1", "b1", "").Return(nil)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerImageRepository.On("BuildImage", configContext.Services[0].DockerImages[0]).Return(nil)

	overrides := domain.NewDockerImageSourceOverrides(map[string]string{"out-of-scope-image": "/local"})

	sut := NewBuildCommandHandler(configRepository, scm, containerImageRepository)

	result := sut.Handle([]string{}, "selected", overrides)

	assert.Nil(t, result)
	scm.AssertExpectations(t)
	containerImageRepository.AssertExpectations(t)
}

func TestBuildCommandHandler_HandleIncludesNonDeployableService(t *testing.T) {
	configContext := &domain.ConfigurationContext{
		Services: []domain.Service{
			{
				Name:         "deployable",
				HelmRepoPath: "any-repo",
				HelmBranch:   "any-branch",
				DockerImages: []domain.DockerImage{
					{
						Name:                     "deployable-image",
						DockerfilePath:           ".",
						BuildContextRelativePath: "",
						GitRepoPath:              "any-repo",
						GitRef:                   "any-branch",
					},
				},
				Profiles: []string{"all"},
			},
			{
				Name: "automation",
				DockerImages: []domain.DockerImage{
					{
						Name:                     "automation-image",
						DockerfilePath:           ".",
						BuildContextRelativePath: "",
						GitRepoPath:              "any-automation-repo",
						GitRef:                   "any-automation-branch",
					},
				},
				Profiles: []string{"all"},
			},
		},
	}
	configRepository := new(testutil.MockConfigRepository)
	configRepository.On("LoadCurrentConfigurationContext").Return(configContext, nil)
	scm := new(testutil.MockScm)
	scm.On("Download", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	containerImageRepository := new(testutil.MockContainerImageRepository)
	containerImageRepository.On("BuildImage", mock.Anything).Return(nil)

	sut := BuildCommandHandler{
		configRepository:         configRepository,
		scm:                      scm,
		containerImageRepository: containerImageRepository,
	}

	result := sut.Handle(nil, "all", domain.DockerImageSourceOverrides{})

	assert.Nil(t, result)
	containerImageRepository.AssertNumberOfCalls(t, "BuildImage", 2)
}
