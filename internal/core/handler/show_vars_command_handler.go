package handler

import (
	"fmt"
	"sort"
	"strings"

	"dx/internal/core"
)

type ShowVarsCommandHandler struct {
	secretsRepository core.SecretsRepository
	configRepository  core.ConfigRepository
}

func ProvideShowVarsCommandHandler(
	secretsRepository core.SecretsRepository,
	configRepository core.ConfigRepository,
) ShowVarsCommandHandler {
	return ShowVarsCommandHandler{
		secretsRepository: secretsRepository,
		configRepository:  configRepository,
	}
}

func (h *ShowVarsCommandHandler) Handle() error {
	values, err := core.CreateTemplatingValues(h.configRepository, h.secretsRepository)
	if err != nil {
		return err
	}

	prettyPrintValuesMap(values)

	return nil
}

func prettyPrintValuesMap(values map[string]interface{}) {
	prettyPrintMap(values, 0, false)
}

func prettyPrintMap(values map[string]interface{}, indent int, hidden bool) {
	// Sort keys for deterministic output
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	indentString := strings.Repeat(" ", indent)
	for _, key := range keys {
		value := values[key]
		if _, ok := value.(string); ok {
			if hidden {
				fmt.Printf("%s%s: ******\n", indentString, key)
			} else {
				fmt.Printf("%s%s: %s\n", indentString, key, value)
			}
		} else {
			fmt.Printf("%s%s:\n", indentString, key)
			if strings.Contains(key, "Secrets") {
				prettyPrintMap(value.(map[string]interface{}), indent+2, true)
			} else {
				prettyPrintMap(value.(map[string]interface{}), indent+2, hidden)
			}
		}
	}
}
