package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustOverrides(t *testing.T, chartDirectoryByService map[string]string) HelmChartOverrides {
	t.Helper()
	overrides, err := NewHelmChartOverrides(chartDirectoryByService)
	require.NoError(t, err)
	return overrides
}

func TestNewHelmChartOverrides_AcceptsValidServiceNames(t *testing.T) {
	_, err := NewHelmChartOverrides(map[string]string{
		"api":         "/a",
		"worker-1":    "/b",
		"frontend_v2": "/c",
	})
	require.NoError(t, err)
}

func TestNewHelmChartOverrides_RejectsInvalidServiceName(t *testing.T) {
	cases := []string{
		"",
		"-leading-dash",
		"a/b",
		`a\b`,
		"a..b",
		"a\x00b",
		"a b",
		"-",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := NewHelmChartOverrides(map[string]string{name: "/path"})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid service name")
		})
	}
}

func TestHelmChartOverrides_LookupChartDirectory_ReturnsConfiguredPath(t *testing.T) {
	overrides := mustOverrides(t, map[string]string{"api": "/override/api"})

	directory, ok := overrides.LookupChartDirectory("api")

	require.True(t, ok)
	assert.Equal(t, "/override/api", directory)
}

func TestHelmChartOverrides_LookupChartDirectory_AbsentForUnconfiguredService(t *testing.T) {
	overrides := mustOverrides(t, map[string]string{"api": "/override/api"})

	_, ok := overrides.LookupChartDirectory("worker")

	assert.False(t, ok)
}

func TestHelmChartOverrides_ValidateAgainstServices_AllowsKnownService(t *testing.T) {
	services := []Service{{Name: "api"}, {Name: "worker"}}
	overrides := mustOverrides(t, map[string]string{"api": "/override/api"})

	require.NoError(t, overrides.ValidateAgainstServices(services))
}

func TestHelmChartOverrides_ValidateAgainstServices_RejectsUnknownService(t *testing.T) {
	services := []Service{{Name: "api"}, {Name: "worker"}}
	overrides := mustOverrides(t, map[string]string{"frontend": "/override/frontend"})

	err := overrides.ValidateAgainstServices(services)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "frontend")
	assert.Contains(t, err.Error(), "api")
	assert.Contains(t, err.Error(), "worker")
	assert.NotContains(t, err.Error(), "--helm-chart", "domain error must not leak the CLI flag name")
}

func TestHelmChartOverrides_ValidateAgainstServices_RejectsMultipleUnknownServices(t *testing.T) {
	services := []Service{{Name: "api"}}
	overrides := mustOverrides(t, map[string]string{
		"frontend": "/override/frontend",
		"worker":   "/override/worker",
	})

	err := overrides.ValidateAgainstServices(services)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "frontend")
	assert.Contains(t, err.Error(), "worker")
}

func TestHelmChartOverrides_ValidateAgainstServices_EmptyIsNoop(t *testing.T) {
	services := []Service{{Name: "api"}}
	overrides := mustOverrides(t, nil)

	require.NoError(t, overrides.ValidateAgainstServices(services))
}

func TestHelmChartOverrides_FindUnusedOverrides_ReturnsNamesNotInScope(t *testing.T) {
	servicesInScope := []Service{{Name: "api"}}
	overrides := mustOverrides(t, map[string]string{
		"api":      "/a",
		"worker":   "/w",
		"frontend": "/f",
	})

	unused := overrides.FindUnusedOverrides(servicesInScope)

	assert.Equal(t, []string{"frontend", "worker"}, unused)
}

func TestHelmChartOverrides_FindUnusedOverrides_EmptyWhenAllInScope(t *testing.T) {
	servicesInScope := []Service{{Name: "api"}, {Name: "worker"}}
	overrides := mustOverrides(t, map[string]string{"api": "/a", "worker": "/w"})

	assert.Empty(t, overrides.FindUnusedOverrides(servicesInScope))
}

func TestHelmChartOverrides_FindUnusedOverrides_NilWhenEmpty(t *testing.T) {
	overrides := mustOverrides(t, nil)
	assert.Nil(t, overrides.FindUnusedOverrides([]Service{{Name: "api"}}))
}

func TestHelmChartOverrides_IsEmpty(t *testing.T) {
	cases := map[string]struct {
		chartDirectoryByService map[string]string
		expectEmpty             bool
	}{
		"nil map":       {nil, true},
		"empty map":     {map[string]string{}, true},
		"non-empty map": {map[string]string{"api": "/override/api"}, false},
	}
	for label, testCase := range cases {
		t.Run(label, func(t *testing.T) {
			assert.Equal(t, testCase.expectEmpty, mustOverrides(t, testCase.chartDirectoryByService).IsEmpty())
		})
	}
}
