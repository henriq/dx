package container_orchestrator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateHelmArgs_AllowedArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "empty args",
			args: []string{},
		},
		{
			name: "set values",
			args: []string{"--set", "foo=bar"},
		},
		{
			name: "values file",
			args: []string{"--values", "values.yaml"},
		},
		{
			name: "multiple allowed flags",
			args: []string{"--set", "image.tag=latest", "--namespace", "production", "--timeout", "5m"},
		},
		{
			name: "set with equals syntax",
			args: []string{"--set=foo=bar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHelmArgs(tt.args)
			assert.NoError(t, err)
		})
	}
}

func TestValidateHelmArgs_BlockedArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		blockedFlag string
	}{
		{
			name:        "post-renderer flag",
			args:        []string{"--post-renderer", "/path/to/script"},
			blockedFlag: "--post-renderer",
		},
		{
			name:        "post-renderer with equals",
			args:        []string{"--post-renderer=/path/to/script"},
			blockedFlag: "--post-renderer",
		},
		{
			name:        "kubeconfig flag",
			args:        []string{"--kubeconfig", "/path/to/config"},
			blockedFlag: "--kubeconfig",
		},
		{
			name:        "kube-context flag",
			args:        []string{"--kube-context", "other-context"},
			blockedFlag: "--kube-context",
		},
		{
			name:        "repository-config flag",
			args:        []string{"--repository-config", "/path/to/repos"},
			blockedFlag: "--repository-config",
		},
		{
			name:        "registry-config flag",
			args:        []string{"--registry-config", "/path/to/registry"},
			blockedFlag: "--registry-config",
		},
		{
			name:        "blocked flag mixed with allowed",
			args:        []string{"--set", "foo=bar", "--post-renderer", "/script"},
			blockedFlag: "--post-renderer",
		},
		{
			name:        "case insensitive check",
			args:        []string{"--POST-RENDERER", "/script"},
			blockedFlag: "--post-renderer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHelmArgs(tt.args)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.blockedFlag)
			assert.Contains(t, err.Error(), "not allowed for security reasons")
		})
	}
}
