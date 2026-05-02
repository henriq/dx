package handler

import (
	"fmt"

	"golang.org/x/mod/semver"

	"pilot/internal/cli/output"
	"pilot/internal/core/domain"
	"pilot/internal/ports"
)

type VersionCheckCommandHandler struct {
	configRepository ports.ConfigRepository
}

func NewVersionCheckCommandHandler(configRepository ports.ConfigRepository) VersionCheckCommandHandler {
	return VersionCheckCommandHandler{
		configRepository: configRepository,
	}
}

// Handle prints a red warning when the running pilot version is older than the configured minimum, and returns whether one was printed.
// Best-effort: config-load errors are swallowed so this never blocks the actual command from running.
func (h *VersionCheckCommandHandler) Handle(currentVersion string) bool {
	if !domain.IsValidPilotVersion(currentVersion) {
		return false
	}

	config, err := h.configRepository.LoadConfig()
	if err != nil {
		return false
	}

	required := config.EffectiveMinPilotVersion()
	if required == "" {
		return false
	}

	if semver.Compare(domain.NormalizePilotVersion(currentVersion), domain.NormalizePilotVersion(required)) >= 0 {
		return false
	}

	message := fmt.Sprintf(
		"pilot %s is older than the minimum required by configuration (%s); upgrade pilot to %s or newer",
		currentVersion,
		required,
		required,
	)
	output.PrintWarningCritical(message)
	return true
}
