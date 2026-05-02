package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func TestConfig_ContextExists(t *testing.T) {
	contextName := string(uuid.NewUUID())
	config := Config{
		Contexts: []ConfigurationContext{
			{
				Name: contextName,
			},
		},
	}
	assert.True(t, config.ContextExists(contextName))
	assert.False(t, config.ContextExists(string(uuid.NewUUID())))
}

func TestConfig_GetContext(t *testing.T) {
	context := ConfigurationContext{
		Name: string(uuid.NewUUID()),
	}
	config := Config{
		Contexts: []ConfigurationContext{context},
	}
	actual, err := config.GetContext(context.Name)
	assert.Nil(t, err)
	assert.Equal(t, context, *actual)
	actual, err = config.GetContext(string(uuid.NewUUID()))
	assert.NotNil(t, err)
	assert.Nil(t, actual)
}

func TestConfigurationContext_GetService(t *testing.T) {
	context := ConfigurationContext{
		Services: []Service{
			{
				Name: string(uuid.NewUUID()),
			},
		},
	}
	actual := context.GetService(context.Services[0].Name)
	assert.Equal(t, context.Services[0], *actual)
	actual = context.GetService(string(uuid.NewUUID()))
	assert.Nil(t, actual)
}

func TestConfigurationContext_FilterServices_ByProfile(t *testing.T) {
	ctx := ConfigurationContext{
		Services: []Service{
			{Name: "svc-a", Profiles: []string{"infra", "all"}},
			{Name: "svc-b", Profiles: []string{"app", "all"}},
			{Name: "svc-c", Profiles: []string{"infra", "all"}},
		},
	}

	result := ctx.FilterServices(nil, "infra")

	assert.Len(t, result, 2)
	assert.Equal(t, "svc-a", result[0].Name)
	assert.Equal(t, "svc-c", result[1].Name)
}

func TestConfigurationContext_FilterServices_ByNames(t *testing.T) {
	ctx := ConfigurationContext{
		Services: []Service{
			{Name: "svc-a", Profiles: []string{"infra"}},
			{Name: "svc-b", Profiles: []string{"app"}},
			{Name: "svc-c", Profiles: []string{"infra"}},
		},
	}

	result := ctx.FilterServices([]string{"svc-b", "svc-c"}, "infra")

	assert.Len(t, result, 2)
	assert.Equal(t, "svc-b", result[0].Name)
	assert.Equal(t, "svc-c", result[1].Name)
}

func TestConfigurationContext_FilterServices_NamesNotFound(t *testing.T) {
	ctx := ConfigurationContext{
		Services: []Service{
			{Name: "svc-a", Profiles: []string{"all"}},
		},
	}

	result := ctx.FilterServices([]string{"nonexistent"}, "all")

	assert.Empty(t, result)
}

func TestConfigurationContext_FilterServices_EmptyServices(t *testing.T) {
	ctx := ConfigurationContext{}

	result := ctx.FilterServices(nil, "all")

	assert.Empty(t, result)
}

func TestConfigurationContext_FilterServices_AllProfile(t *testing.T) {
	ctx := ConfigurationContext{
		Services: []Service{
			{Name: "svc-a", Profiles: []string{"infra", "all"}},
			{Name: "svc-b", Profiles: []string{"app", "all"}},
		},
	}

	result := ctx.FilterServices(nil, "all")

	assert.Len(t, result, 2)
}

func TestCreateDefaultConfigReturnsConfig(t *testing.T) {
	defaultConfig := CreateDefaultConfig()
	assert.NotNil(t, defaultConfig)
	assert.Equal(t, 1, len(defaultConfig.Contexts))
	assert.Nil(t, defaultConfig.Validate())
}

func TestConfig_Validate_ContextNamePathTraversal(t *testing.T) {
	tests := []struct {
		name        string
		contextName string
		wantErr     bool
	}{
		{"valid name", "my-context", false},
		{"valid name with dashes", "my-context-123", false},
		{"path traversal with ..", "../etc", true},
		{"path traversal with forward slash", "foo/bar", true},
		{"path traversal with backslash", "foo\\bar", true},
		{"null byte injection", "foo\x00bar", true},
		{"double dot in middle", "foo..bar", true},
		{"empty name", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Contexts: []ConfigurationContext{
					{
						Name: tt.contextName,
						Services: []Service{
							{
								Name:                  "test-svc",
								HelmRepoPath:          "/tmp/helm",
								HelmBranch:            "main",
								HelmChartRelativePath: "charts",
								DockerImages: []DockerImage{
									{
										Name:                     "test-img",
										DockerfilePath:           "Dockerfile",
										BuildContextRelativePath: ".",
										GitRepoPath:              "/tmp/repo",
										GitRef:                   "main",
									},
								},
							},
						},
					},
				},
			}

			err := config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNormalizePilotVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"v1.2.3", "v1.2.3"},
		{"1.2.3", "v1.2.3"},
		{"  1.2.3  ", "v1.2.3"},
		{"", ""},
		{"v0.0.0-prerelease+abc", "v0.0.0-prerelease+abc"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, NormalizePilotVersion(tt.input), tt.input)
	}
}

func TestIsValidPilotVersion(t *testing.T) {
	assert.True(t, IsValidPilotVersion("v1.2.3"))
	assert.True(t, IsValidPilotVersion("1.2.3"))
	assert.True(t, IsValidPilotVersion("v1.2.3-prerelease+abc"))
	assert.False(t, IsValidPilotVersion("dev"))
	assert.False(t, IsValidPilotVersion("not-a-version"))
	assert.False(t, IsValidPilotVersion(""))
}

func TestHigherPilotVersion(t *testing.T) {
	tests := []struct {
		a, b, expected string
	}{
		{"", "", ""},
		{"v1.2.3", "", "v1.2.3"},
		{"", "v1.2.3", "v1.2.3"},
		{"v1.2.3", "v1.2.4", "v1.2.4"},
		{"v2.0.0", "v1.9.9", "v2.0.0"},
		{"1.2.3", "v1.2.4", "v1.2.4"},
		{"v1.2.3", "v1.2.3", "v1.2.3"},
		{"v1.2.3-prerelease", "v1.2.3", "v1.2.3"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, HigherPilotVersion(tt.a, tt.b), "a=%s b=%s", tt.a, tt.b)
	}
}

func TestConfig_EffectiveMinPilotVersion(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected string
	}{
		{
			name:     "all empty",
			config:   Config{Contexts: []ConfigurationContext{{Name: "a"}, {Name: "b"}}},
			expected: "",
		},
		{
			name: "root only",
			config: Config{
				MinPilotVersion: "v1.2.3",
				Contexts:        []ConfigurationContext{{Name: "a"}},
			},
			expected: "v1.2.3",
		},
		{
			name: "context only",
			config: Config{
				Contexts: []ConfigurationContext{{Name: "a", MinPilotVersion: "v1.5.0"}},
			},
			expected: "v1.5.0",
		},
		{
			name: "context wins over root",
			config: Config{
				MinPilotVersion: "v1.0.0",
				Contexts: []ConfigurationContext{
					{Name: "a", MinPilotVersion: "v0.9.0"},
					{Name: "b", MinPilotVersion: "v2.0.0"},
				},
			},
			expected: "v2.0.0",
		},
		{
			name: "root wins over contexts",
			config: Config{
				MinPilotVersion: "v3.0.0",
				Contexts: []ConfigurationContext{
					{Name: "a", MinPilotVersion: "v1.5.0"},
					{Name: "b", MinPilotVersion: "v2.0.0"},
				},
			},
			expected: "v3.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.EffectiveMinPilotVersion())
		})
	}
}

func TestConfig_Validate_MinPilotVersion(t *testing.T) {
	baseService := Service{
		Name:                  "test-svc",
		HelmRepoPath:          "/tmp/helm",
		HelmBranch:            "main",
		HelmChartRelativePath: "charts",
		DockerImages: []DockerImage{
			{
				Name:                     "test-img",
				DockerfilePath:           "Dockerfile",
				BuildContextRelativePath: ".",
				GitRepoPath:              "/tmp/repo",
				GitRef:                   "main",
			},
		},
	}

	tests := []struct {
		name           string
		rootMinVersion string
		ctxMinVersion  string
		wantErr        bool
	}{
		{name: "both empty", wantErr: false},
		{name: "root valid", rootMinVersion: "v1.2.3", wantErr: false},
		{name: "root valid no prefix", rootMinVersion: "1.2.3", wantErr: false},
		{name: "ctx valid", ctxMinVersion: "v1.2.3", wantErr: false},
		{name: "root invalid", rootMinVersion: "not-a-version", wantErr: true},
		{name: "ctx invalid", ctxMinVersion: "1..2.3", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				MinPilotVersion: tt.rootMinVersion,
				Contexts: []ConfigurationContext{
					{
						Name:            "ctx",
						MinPilotVersion: tt.ctxMinVersion,
						Services:        []Service{baseService},
					},
				},
			}
			err := config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
