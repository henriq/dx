package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustImageSourceOverrides(t *testing.T, sourcePathByImage map[string]string) DockerImageSourceOverrides {
	t.Helper()
	return NewDockerImageSourceOverrides(sourcePathByImage)
}

func TestDockerImageSourceOverrides_LookupSourcePath_ReturnsConfiguredPath(t *testing.T) {
	overrides := mustImageSourceOverrides(t, map[string]string{"api": "/override/api"})

	path, ok := overrides.LookupSourcePath("api")

	require.True(t, ok)
	assert.Equal(t, "/override/api", path)
}

func TestDockerImageSourceOverrides_LookupSourcePath_AbsentForUnconfiguredImage(t *testing.T) {
	overrides := mustImageSourceOverrides(t, map[string]string{"api": "/override/api"})

	_, ok := overrides.LookupSourcePath("worker")

	assert.False(t, ok)
}

func TestDockerImageSourceOverrides_ValidateAgainstImages_AllowsKnownImage(t *testing.T) {
	images := []DockerImage{{Name: "api"}, {Name: "worker"}}
	overrides := mustImageSourceOverrides(t, map[string]string{"api": "/override/api"})

	require.NoError(t, overrides.ValidateAgainstImages(images))
}

func TestDockerImageSourceOverrides_ValidateAgainstImages_RejectsUnknownImage(t *testing.T) {
	images := []DockerImage{{Name: "api"}, {Name: "worker"}}
	overrides := mustImageSourceOverrides(t, map[string]string{"frontend": "/override/frontend"})

	err := overrides.ValidateAgainstImages(images)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "frontend")
	assert.Contains(t, err.Error(), "api")
	assert.Contains(t, err.Error(), "worker")
	assert.NotContains(t, err.Error(), "--image-source", "domain error must not leak the CLI flag name")
}

func TestDockerImageSourceOverrides_ValidateAgainstImages_ListsAvailableImagesAsSortedBullets(t *testing.T) {
	images := []DockerImage{{Name: "worker"}, {Name: "api"}, {Name: "frontend"}}
	overrides := mustImageSourceOverrides(t, map[string]string{"missing": "/override/missing"})

	err := overrides.ValidateAgainstImages(images)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "available images:\n  - api\n  - frontend\n  - worker")
}

func TestDockerImageSourceOverrides_ValidateAgainstImages_RejectsMultipleUnknownImages(t *testing.T) {
	images := []DockerImage{{Name: "api"}}
	overrides := mustImageSourceOverrides(t, map[string]string{
		"frontend": "/override/frontend",
		"worker":   "/override/worker",
	})

	err := overrides.ValidateAgainstImages(images)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "frontend")
	assert.Contains(t, err.Error(), "worker")
}

func TestDockerImageSourceOverrides_ValidateAgainstImages_EmptyIsNoop(t *testing.T) {
	images := []DockerImage{{Name: "api"}}
	overrides := mustImageSourceOverrides(t, nil)

	require.NoError(t, overrides.ValidateAgainstImages(images))
}

func TestDockerImageSourceOverrides_FindUnusedOverrides_ReturnsNamesNotInScope(t *testing.T) {
	imagesInScope := []DockerImage{{Name: "api"}}
	overrides := mustImageSourceOverrides(t, map[string]string{
		"api":      "/a",
		"worker":   "/w",
		"frontend": "/f",
	})

	unused := overrides.FindUnusedOverrides(imagesInScope)

	assert.Equal(t, []string{"frontend", "worker"}, unused)
}

func TestDockerImageSourceOverrides_FindUnusedOverrides_EmptyWhenAllInScope(t *testing.T) {
	imagesInScope := []DockerImage{{Name: "api"}, {Name: "worker"}}
	overrides := mustImageSourceOverrides(t, map[string]string{"api": "/a", "worker": "/w"})

	assert.Empty(t, overrides.FindUnusedOverrides(imagesInScope))
}

func TestDockerImageSourceOverrides_FindUnusedOverrides_NilWhenEmpty(t *testing.T) {
	overrides := mustImageSourceOverrides(t, nil)
	assert.Nil(t, overrides.FindUnusedOverrides([]DockerImage{{Name: "api"}}))
}

func TestDockerImageSourceOverrides_IsEmpty(t *testing.T) {
	cases := map[string]struct {
		sourcePathByImage map[string]string
		expectEmpty       bool
	}{
		"nil map":       {nil, true},
		"empty map":     {map[string]string{}, true},
		"non-empty map": {map[string]string{"api": "/override/api"}, false},
	}
	for label, testCase := range cases {
		t.Run(label, func(t *testing.T) {
			assert.Equal(t, testCase.expectEmpty, mustImageSourceOverrides(t, testCase.sourcePathByImage).IsEmpty())
		})
	}
}
