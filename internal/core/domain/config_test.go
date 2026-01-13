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
