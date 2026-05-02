package handler

import (
	"errors"
	"testing"

	"pilot/internal/core/domain"
	"pilot/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func TestVersionCheckCommandHandler_Handle(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		config         *domain.Config
		loadErr        error
		wantWarn       bool
	}{
		{
			name:           "no config — no warning",
			currentVersion: "v1.0.0",
			loadErr:        errors.New("config missing"),
			wantWarn:       false,
		},
		{
			name:           "no minimum required — no warning",
			currentVersion: "v1.0.0",
			config:         &domain.Config{Contexts: []domain.ConfigurationContext{{Name: "a"}}},
			wantWarn:       false,
		},
		{
			name:           "current equal to minimum — no warning",
			currentVersion: "v1.2.3",
			config: &domain.Config{
				MinPilotVersion: "v1.2.3",
				Contexts:        []domain.ConfigurationContext{{Name: "a"}},
			},
			wantWarn: false,
		},
		{
			name:           "current newer than minimum — no warning",
			currentVersion: "v2.0.0",
			config: &domain.Config{
				MinPilotVersion: "v1.2.3",
				Contexts:        []domain.ConfigurationContext{{Name: "a"}},
			},
			wantWarn: false,
		},
		{
			name:           "current older than root minimum — warning",
			currentVersion: "v1.0.0",
			config: &domain.Config{
				MinPilotVersion: "v1.5.0",
				Contexts:        []domain.ConfigurationContext{{Name: "a"}},
			},
			wantWarn: true,
		},
		{
			name:           "current older than context minimum — warning",
			currentVersion: "v1.0.0",
			config: &domain.Config{
				Contexts: []domain.ConfigurationContext{
					{Name: "a", MinPilotVersion: "v1.5.0"},
				},
			},
			wantWarn: true,
		},
		{
			name:           "highest of root and contexts wins",
			currentVersion: "v2.0.0",
			config: &domain.Config{
				MinPilotVersion: "v1.0.0",
				Contexts: []domain.ConfigurationContext{
					{Name: "a", MinPilotVersion: "v3.0.0"},
				},
			},
			wantWarn: true,
		},
		{
			name:           "dev build — never warns",
			currentVersion: "dev",
			config: &domain.Config{
				MinPilotVersion: "v999.0.0",
				Contexts:        []domain.ConfigurationContext{{Name: "a"}},
			},
			wantWarn: false,
		},
		{
			name:           "current accepts no v prefix",
			currentVersion: "1.0.0",
			config: &domain.Config{
				MinPilotVersion: "v1.5.0",
				Contexts:        []domain.ConfigurationContext{{Name: "a"}},
			},
			wantWarn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configRepo := new(testutil.MockConfigRepository)
			if tt.loadErr != nil {
				configRepo.On("LoadConfig").Return(nil, tt.loadErr)
			} else if tt.config != nil {
				configRepo.On("LoadConfig").Return(tt.config, nil)
			}

			sut := NewVersionCheckCommandHandler(configRepo)
			warned := sut.Handle(tt.currentVersion)

			assert.Equal(t, tt.wantWarn, warned)
		})
	}
}
