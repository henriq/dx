package core

import (
	"testing"

	"pilot/internal/core/domain"

	"github.com/stretchr/testify/assert"
)

func TestExtractTemplateVariables_SecretsOnly(t *testing.T) {
	template := `--set=password={{.Secrets.DB_PASSWORD}}`

	result := ExtractTemplateVariables(template)

	assert.Equal(t, []string{"DB_PASSWORD"}, result["Secrets"])
	assert.Empty(t, result["Services"])
}

func TestExtractTemplateVariables_MultipleSecrets(t *testing.T) {
	template := `--set=password={{.Secrets.DB_PASSWORD}} --set=apiKey={{.Secrets.API_KEY}}`

	result := ExtractTemplateVariables(template)

	assert.ElementsMatch(t, []string{"DB_PASSWORD", "API_KEY"}, result["Secrets"])
}

func TestExtractTemplateVariables_DuplicateSecrets(t *testing.T) {
	template := `{{.Secrets.DB_PASSWORD}} and {{.Secrets.DB_PASSWORD}} again`

	result := ExtractTemplateVariables(template)

	assert.Equal(t, []string{"DB_PASSWORD"}, result["Secrets"])
}

func TestExtractTemplateVariables_ServicesOnly(t *testing.T) {
	template := `cd {{.Services.api.path}} && make build`

	result := ExtractTemplateVariables(template)

	assert.Equal(t, []string{"api"}, result["Services"])
	assert.Empty(t, result["Secrets"])
}

func TestExtractTemplateVariables_MixedTypes(t *testing.T) {
	template := `cd {{.Services.api.path}} && echo {{.Secrets.TOKEN}}`

	result := ExtractTemplateVariables(template)

	assert.Equal(t, []string{"api"}, result["Services"])
	assert.Equal(t, []string{"TOKEN"}, result["Secrets"])
}

func TestExtractTemplateVariables_WithWhitespaceTrim(t *testing.T) {
	template := `{{- .Secrets.DB_PASSWORD -}}`

	result := ExtractTemplateVariables(template)

	assert.Equal(t, []string{"DB_PASSWORD"}, result["Secrets"])
}

func TestExtractTemplateVariables_EmptyTemplate(t *testing.T) {
	result := ExtractTemplateVariables("")

	assert.Empty(t, result)
}

func TestExtractTemplateVariables_NoVariables(t *testing.T) {
	template := `just some plain text without any template variables`

	result := ExtractTemplateVariables(template)

	assert.Empty(t, result)
}

func TestExtractSecretKeys_FromScripts(t *testing.T) {
	ctx := &domain.ConfigurationContext{
		Scripts: map[string]string{
			"deploy":  "echo {{.Secrets.DEPLOY_TOKEN}}",
			"migrate": "DATABASE_URL={{.Secrets.DB_URL}} ./migrate",
		},
	}

	keys := ExtractSecretKeys(ctx)

	assert.ElementsMatch(t, []string{"DB_URL", "DEPLOY_TOKEN"}, keys)
}

func TestExtractSecretKeys_FromHelmArgs(t *testing.T) {
	ctx := &domain.ConfigurationContext{
		Services: []domain.Service{
			{
				Name: "api",
				HelmArgs: []string{
					"--set=password={{.Secrets.DB_PASSWORD}}",
					"--set=apiKey={{.Secrets.API_KEY}}",
				},
			},
		},
	}

	keys := ExtractSecretKeys(ctx)

	assert.ElementsMatch(t, []string{"API_KEY", "DB_PASSWORD"}, keys)
}

func TestExtractSecretKeys_FromBuildArgs(t *testing.T) {
	ctx := &domain.ConfigurationContext{
		Services: []domain.Service{
			{
				Name: "api",
				DockerImages: []domain.DockerImage{
					{
						Name: "api-image",
						BuildArgs: []string{
							"--build-arg=NPM_TOKEN={{.Secrets.NPM_TOKEN}}",
						},
					},
				},
			},
		},
	}

	keys := ExtractSecretKeys(ctx)

	assert.Equal(t, []string{"NPM_TOKEN"}, keys)
}

func TestExtractSecretKeys_AllSources(t *testing.T) {
	ctx := &domain.ConfigurationContext{
		Scripts: map[string]string{
			"test": "TOKEN={{.Secrets.SCRIPT_SECRET}} ./test.sh",
		},
		Services: []domain.Service{
			{
				Name: "api",
				HelmArgs: []string{
					"--set=dbPass={{.Secrets.HELM_SECRET}}",
				},
				DockerImages: []domain.DockerImage{
					{
						Name: "api-image",
						BuildArgs: []string{
							"--build-arg=TOKEN={{.Secrets.BUILD_SECRET}}",
						},
					},
				},
			},
		},
	}

	keys := ExtractSecretKeys(ctx)

	assert.ElementsMatch(t, []string{"BUILD_SECRET", "HELM_SECRET", "SCRIPT_SECRET"}, keys)
}

func TestExtractSecretKeys_Deduplicated(t *testing.T) {
	ctx := &domain.ConfigurationContext{
		Scripts: map[string]string{
			"script1": "{{.Secrets.SHARED_SECRET}}",
			"script2": "{{.Secrets.SHARED_SECRET}}",
		},
		Services: []domain.Service{
			{
				Name:     "api",
				HelmArgs: []string{"--set=s={{.Secrets.SHARED_SECRET}}"},
			},
		},
	}

	keys := ExtractSecretKeys(ctx)

	assert.Equal(t, []string{"SHARED_SECRET"}, keys)
}

func TestExtractSecretKeys_EmptyContext(t *testing.T) {
	ctx := &domain.ConfigurationContext{}

	keys := ExtractSecretKeys(ctx)

	assert.Empty(t, keys)
}

func TestExtractSecretKeys_NoSecrets(t *testing.T) {
	ctx := &domain.ConfigurationContext{
		Scripts: map[string]string{
			"test": "echo hello",
		},
		Services: []domain.Service{
			{
				Name:     "api",
				HelmArgs: []string{"--set=replicas=3"},
			},
		},
	}

	keys := ExtractSecretKeys(ctx)

	assert.Empty(t, keys)
}

func TestExtractSecretKeys_Sorted(t *testing.T) {
	ctx := &domain.ConfigurationContext{
		Scripts: map[string]string{
			"a": "{{.Secrets.ZEBRA}}",
			"b": "{{.Secrets.ALPHA}}",
			"c": "{{.Secrets.MIKE}}",
		},
	}

	keys := ExtractSecretKeys(ctx)

	assert.Equal(t, []string{"ALPHA", "MIKE", "ZEBRA"}, keys)
}

func TestExtractServiceReferences_SingleService(t *testing.T) {
	template := `cd {{.Services.api.path}} && make build`

	refs := ExtractServiceReferences(template)

	assert.Equal(t, []string{"api"}, refs)
}

func TestExtractServiceReferences_MultipleServices(t *testing.T) {
	template := `{{.Services.api.path}} and {{.Services.frontend.path}}`

	refs := ExtractServiceReferences(template)

	assert.ElementsMatch(t, []string{"api", "frontend"}, refs)
}

func TestExtractServiceReferences_NoServices(t *testing.T) {
	template := `echo {{.Secrets.TOKEN}}`

	refs := ExtractServiceReferences(template)

	assert.Empty(t, refs)
}

func TestExtractServiceReferences_HyphenatedServiceName(t *testing.T) {
	template := `cd {{ (index .Services "my-service").path }} && make build`

	refs := ExtractServiceReferences(template)

	assert.Equal(t, []string{"my-service"}, refs)
}

func TestExtractTemplateVariables_HyphenatedKey(t *testing.T) {
	template := `{{ (index .Services "my-service").path }}`

	result := ExtractTemplateVariables(template)

	assert.Equal(t, []string{"my-service"}, result["Services"])
}

// Nested secret key tests - secret keys can contain dots (e.g., "foo.bar" is the key)
func TestExtractTemplateVariables_NestedSecretKey(t *testing.T) {
	template := `{{.Secrets.foo.bar}}`

	result := ExtractTemplateVariables(template)

	// The full key "foo.bar" should be extracted, not just "foo"
	assert.Equal(t, []string{"foo.bar"}, result["Secrets"])
}

func TestExtractTemplateVariables_DeeplyNestedSecretKey(t *testing.T) {
	template := `{{.Secrets.database.credentials.password}}`

	result := ExtractTemplateVariables(template)

	assert.Equal(t, []string{"database.credentials.password"}, result["Secrets"])
}

func TestExtractTemplateVariables_MixedNestedAndSimpleSecrets(t *testing.T) {
	template := `{{.Secrets.simple}} and {{.Secrets.nested.key}}`

	result := ExtractTemplateVariables(template)

	assert.ElementsMatch(t, []string{"simple", "nested.key"}, result["Secrets"])
}

func TestExtractSecretKeys_NestedSecretKey(t *testing.T) {
	ctx := &domain.ConfigurationContext{
		Scripts: map[string]string{
			"deploy": "echo {{.Secrets.database.password}}",
		},
	}

	keys := ExtractSecretKeys(ctx)

	assert.Equal(t, []string{"database.password"}, keys)
}

func TestExtractSecretKeys_NestedAndSimpleSecrets(t *testing.T) {
	ctx := &domain.ConfigurationContext{
		Services: []domain.Service{
			{
				Name: "api",
				HelmArgs: []string{
					"--set=dbPass={{.Secrets.db.password}}",
					"--set=apiKey={{.Secrets.API_KEY}}",
				},
			},
		},
	}

	keys := ExtractSecretKeys(ctx)

	assert.ElementsMatch(t, []string{"API_KEY", "db.password"}, keys)
}

// Services should still only extract the first part (service name)
func TestExtractTemplateVariables_ServiceWithNestedProperty(t *testing.T) {
	template := `{{.Services.api.config.port}}`

	result := ExtractTemplateVariables(template)

	// For services, only extract the service name "api", not the full path
	assert.Equal(t, []string{"api"}, result["Services"])
}

// Index function syntax tests - for keys with special characters like hyphens
func TestExtractTemplateVariables_IndexFunctionService(t *testing.T) {
	template := `echo {{ (index .Services "my-service").path }}`

	result := ExtractTemplateVariables(template)

	assert.Equal(t, []string{"my-service"}, result["Services"])
}

func TestExtractTemplateVariables_IndexFunctionSecret(t *testing.T) {
	template := `--set=password={{ (index .Secrets "db-password") }}`

	result := ExtractTemplateVariables(template)

	assert.Equal(t, []string{"db-password"}, result["Secrets"])
}

func TestExtractTemplateVariables_IndexFunctionWithWhitespaceTrim(t *testing.T) {
	template := `{{- (index .Services "test-service").path -}}`

	result := ExtractTemplateVariables(template)

	assert.Equal(t, []string{"test-service"}, result["Services"])
}

func TestExtractTemplateVariables_MixedSyntax(t *testing.T) {
	template := `cd {{.Services.api.path}} && echo {{ (index .Services "other-service").url }}`

	result := ExtractTemplateVariables(template)

	assert.ElementsMatch(t, []string{"api", "other-service"}, result["Services"])
}

func TestExtractServiceReferences_IndexFunction(t *testing.T) {
	template := `cd {{ (index .Services "my-service").path }} && make build`

	refs := ExtractServiceReferences(template)

	assert.Equal(t, []string{"my-service"}, refs)
}

// Pipeline syntax
func TestExtractTemplateVariables_WithPipeline(t *testing.T) {
	template := `{{ .Secrets.KEY | print }}`
	result := ExtractTemplateVariables(template)
	assert.Equal(t, []string{"KEY"}, result["Secrets"])
}

func TestExtractTemplateVariables_WithMultiplePipelines(t *testing.T) {
	template := `{{ .Secrets.DB_PASSWORD | printf "%s" | print }}`
	result := ExtractTemplateVariables(template)
	assert.Equal(t, []string{"DB_PASSWORD"}, result["Secrets"])
}

func TestExtractTemplateVariables_ServiceWithPipeline(t *testing.T) {
	template := `{{ .Services.api.path | print }}`
	result := ExtractTemplateVariables(template)
	assert.Equal(t, []string{"api"}, result["Services"])
}

// Trim markers
func TestExtractTemplateVariables_TrailingTrimMarker(t *testing.T) {
	template := `{{.Secrets.KEY -}}`
	result := ExtractTemplateVariables(template)
	assert.Equal(t, []string{"KEY"}, result["Secrets"])
}

func TestExtractTemplateVariables_BothTrimMarkers(t *testing.T) {
	template := `{{- .Secrets.KEY -}}`
	result := ExtractTemplateVariables(template)
	assert.Equal(t, []string{"KEY"}, result["Secrets"])
}

// Index function with backticks
func TestExtractTemplateVariables_IndexFunctionBackticks(t *testing.T) {
	template := "{{ (index .Secrets `db-password`) }}"
	result := ExtractTemplateVariables(template)
	assert.Equal(t, []string{"db-password"}, result["Secrets"])
}

// Nil context
func TestExtractSecretKeys_NilContext(t *testing.T) {
	keys := ExtractSecretKeys(nil)
	assert.Nil(t, keys)
}

// Sorted output
func TestExtractTemplateVariables_SortedKeys(t *testing.T) {
	template := `{{.Secrets.ZEBRA}} {{.Secrets.ALPHA}} {{.Secrets.MIKE}}`
	result := ExtractTemplateVariables(template)
	assert.Equal(t, []string{"ALPHA", "MIKE", "ZEBRA"}, result["Secrets"])
}

func TestExtractTemplateVariables_SortedServices(t *testing.T) {
	template := `{{.Services.zebra.path}} {{.Services.alpha.path}} {{.Services.mike.path}}`
	result := ExtractTemplateVariables(template)
	assert.Equal(t, []string{"alpha", "mike", "zebra"}, result["Services"])
}

func TestReferencesAllServices(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     bool
	}{
		{"range over services with vars", `{{ range $name, $service := .Services }}{{ $name }}{{ end }}`, true},
		{"range without vars", `{{ range .Services }}x{{ end }}`, true},
		{"with block", `{{ with .Services }}{{ len . }}{{ end }}`, true},
		{"len call", `{{ if eq (len .Services) 0 }}empty{{ end }}`, true},
		{"plain reference", `{{ .Services }}`, true},
		{"specific service via dot", `cd {{ .Services.api.path }} && make build`, false},
		{"index function", `cd {{ (index .Services "my-service").path }} && make build`, false},
		{"longer identifier", `{{ .ServicesList }}`, false},
		{"no template", `echo hello`, false},
		{"mixed root and index", `{{ range .Services }}{{ end }}{{ (index .Services "x").path }}`, true},
		{"only index references", `{{ (index .Services "a").path }} {{ (index .Services "b").path }}`, false},
		{"empty string", ``, false},
		{"multi-line range", "echo before\n{{ range $name, $service := .Services }}\n{{ $name }}\n{{ end }}", true},
		{"trim markers", `{{- .Services -}}`, true},
		{"assignment to variable", `{{ $svc := .Services }}{{ range $svc }}x{{ end }}`, true},
		{"if pipeline", `{{ if .Services }}has services{{ end }}`, true},
		{"services-map prefix", `{{ .ServicesMap.foo }}`, false},
		{"inside template comment", `{{/* placeholder for .Services iteration */}}`, false},
		{"dollar root scope", `{{ range .Services }}{{ $.Services }}{{ end }}`, true},
		{"template invocation", `{{ template "x" .Services }}`, true},
		{"chained access on parenthesized root", `{{ (.Services).foo }}`, true},
		{"pipeline into len", `{{ .Services | len }}`, true},
		{"index with dynamic key", `{{ (index .Services .Key) }}`, true},
		{"multi-arg index with string keys", `{{ index .Services "a" "nested" }}`, false},
		{"malformed template", `{{ .Services`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ReferencesAllServices(tt.template))
		})
	}
}

func TestExtractTemplateVariables_MalformedTemplate(t *testing.T) {
	template := `{{ .Services`
	assert.Empty(t, ExtractTemplateVariables(template))
}

func TestExtractServiceReferences_MalformedTemplate(t *testing.T) {
	template := `{{ .Services`
	assert.Empty(t, ExtractServiceReferences(template))
}

func TestExtractServiceReferences_QuotedDotKey_NoLongerSupported(t *testing.T) {
	template := `{{ .Services."my-service".path }}`
	assert.Empty(t, ExtractServiceReferences(template))
}

func TestExtractServiceReferences_BareIndexCall(t *testing.T) {
	template := `{{ index .Services "my-service" }}`
	assert.Equal(t, []string{"my-service"}, ExtractServiceReferences(template))
	assert.False(t, ReferencesAllServices(template))
}

func TestExtractServiceReferences_BareIndexCallWithPipeline(t *testing.T) {
	template := `{{ index .Services "my-service" | print }}`
	assert.Equal(t, []string{"my-service"}, ExtractServiceReferences(template))
	assert.False(t, ReferencesAllServices(template))
}

func TestExtractTemplateVariables_IndexFunctionDottedSecretKey(t *testing.T) {
	template := `{{ (index .Secrets "db.password") }}`
	result := ExtractTemplateVariables(template)
	assert.Equal(t, []string{"db.password"}, result["Secrets"])
}

func TestExtractServiceReferences_MultiArgIndex(t *testing.T) {
	template := `{{ index .Services "a" "nested" }}`
	assert.Equal(t, []string{"a"}, ExtractServiceReferences(template))
	assert.False(t, ReferencesAllServices(template))
}

func TestReferencesAllServices_RangeBodyFieldAccess(t *testing.T) {
	template := `{{ range .Services }}{{ .gitRef }}{{ end }}`
	assert.True(t, ReferencesAllServices(template))
	assert.Empty(t, ExtractServiceReferences(template))
}
