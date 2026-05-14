package cmd

import (
	"testing"

	"pilot/internal/core/domain"

	"github.com/stretchr/testify/assert"
)

func TestFormatServiceProfiles(t *testing.T) {
	tests := []struct {
		name    string
		service domain.Service
		want    string
	}{
		{
			name: "deployable with profiles returns sorted profiles",
			service: domain.Service{
				HelmRepoPath:          "r",
				HelmBranch:            "b",
				HelmChartRelativePath: "c",
				Profiles:              []string{"staging", "default"},
			},
			want: "default, staging",
		},
		{
			name: "deployable with no profiles returns empty string",
			service: domain.Service{
				HelmRepoPath:          "r",
				HelmBranch:            "b",
				HelmChartRelativePath: "c",
			},
			want: "",
		},
		{
			name: "non-deployable with profiles returns profiles plus suffix",
			service: domain.Service{
				Profiles: []string{"all"},
			},
			want: "all (non-deployable)",
		},
		{
			name:    "non-deployable with no profiles returns only suffix",
			service: domain.Service{},
			want:    "(non-deployable)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, formatServiceProfiles(tt.service))
		})
	}
}
