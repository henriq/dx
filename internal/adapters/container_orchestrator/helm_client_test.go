package container_orchestrator

import (
	"errors"
	"io"
	"testing"

	"dx/internal/ports"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCommandRunner implements ports.CommandRunner for testing
type mockCommandRunner struct {
	runFunc          func(name string, args ...string) ([]byte, error)
	runWithEnvFunc   func(name string, env []string, args ...string) ([]byte, error)
	runInDirFunc     func(dir, name string, args ...string) ([]byte, error)
	runWithStdinFunc func(stdin io.Reader, name string, args ...string) ([]byte, error)
}

func (m *mockCommandRunner) Run(name string, args ...string) ([]byte, error) {
	if m.runFunc != nil {
		return m.runFunc(name, args...)
	}
	return nil, nil
}

func (m *mockCommandRunner) RunWithEnv(name string, env []string, args ...string) ([]byte, error) {
	if m.runWithEnvFunc != nil {
		return m.runWithEnvFunc(name, env, args...)
	}
	return nil, nil
}

func (m *mockCommandRunner) RunInDir(dir, name string, args ...string) ([]byte, error) {
	if m.runInDirFunc != nil {
		return m.runInDirFunc(dir, name, args...)
	}
	return nil, nil
}

func (m *mockCommandRunner) RunWithStdin(stdin io.Reader, name string, args ...string) ([]byte, error) {
	if m.runWithStdinFunc != nil {
		return m.runWithStdinFunc(stdin, name, args...)
	}
	return nil, nil
}

func (m *mockCommandRunner) RunInteractive(name string, args ...string) error {
	return nil
}

func TestHelmClient_Template(t *testing.T) {
	var capturedArgs []string
	runner := &mockCommandRunner{
		runFunc: func(name string, args ...string) ([]byte, error) {
			if name == "helm" && len(args) > 0 && args[0] == "template" {
				capturedArgs = args
				return []byte("apiVersion: v1\nkind: ConfigMap\n"), nil
			}
			return nil, nil
		},
	}

	client := ProvideHelmClient(runner)

	output, err := client.Template("my-release", "/path/to/chart", "my-namespace", []string{"--set", "foo=bar"})
	require.NoError(t, err)
	assert.Contains(t, string(output), "ConfigMap")
	assert.Equal(t, []string{"template", "my-release", "/path/to/chart", "--namespace", "my-namespace", "--set", "foo=bar"}, capturedArgs)
}

func TestHelmClient_Template_NoNamespace(t *testing.T) {
	var capturedArgs []string
	runner := &mockCommandRunner{
		runFunc: func(name string, args ...string) ([]byte, error) {
			if name == "helm" && len(args) > 0 && args[0] == "template" {
				capturedArgs = args
				return []byte(""), nil
			}
			return nil, nil
		},
	}

	client := ProvideHelmClient(runner)

	_, err := client.Template("my-release", "/path/to/chart", "", nil)
	require.NoError(t, err)
	// Should not include --namespace when empty
	assert.Equal(t, []string{"template", "my-release", "/path/to/chart"}, capturedArgs)
}

func TestHelmClient_Template_Error(t *testing.T) {
	runner := &mockCommandRunner{
		runFunc: func(name string, args ...string) ([]byte, error) {
			return []byte("Error: chart not found"), errors.New("exit status 1")
		},
	}

	client := ProvideHelmClient(runner)

	_, err := client.Template("my-release", "/path/to/chart", "", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "helm template failed")
}

func TestHelmClient_UpgradeFromManifests(t *testing.T) {
	var capturedArgs []string
	runner := &mockCommandRunner{
		runFunc: func(name string, args ...string) ([]byte, error) {
			if name == "helm" && len(args) > 0 && args[0] == "upgrade" {
				capturedArgs = args
				return []byte("Release \"my-release\" has been upgraded."), nil
			}
			return nil, nil
		},
	}

	client := ProvideHelmClient(runner)

	err := client.UpgradeFromManifests("my-release", "my-namespace", "/path/to/wrapper")
	require.NoError(t, err)
	assert.Equal(t, []string{"upgrade", "--install", "--labels", "managed-by=dx", "my-release", "/path/to/wrapper", "--namespace", "my-namespace"}, capturedArgs)
}

func TestHelmClient_UpgradeFromManifests_NoNamespace(t *testing.T) {
	var capturedArgs []string
	runner := &mockCommandRunner{
		runFunc: func(name string, args ...string) ([]byte, error) {
			if name == "helm" && len(args) > 0 && args[0] == "upgrade" {
				capturedArgs = args
				return []byte(""), nil
			}
			return nil, nil
		},
	}

	client := ProvideHelmClient(runner)

	err := client.UpgradeFromManifests("my-release", "", "/path/to/wrapper")
	require.NoError(t, err)
	assert.Equal(t, []string{"upgrade", "--install", "--labels", "managed-by=dx", "my-release", "/path/to/wrapper"}, capturedArgs)
}

func TestHelmClient_Uninstall(t *testing.T) {
	var capturedArgs []string
	runner := &mockCommandRunner{
		runFunc: func(name string, args ...string) ([]byte, error) {
			if name == "helm" && len(args) > 0 && args[0] == "uninstall" {
				capturedArgs = args
				return []byte(""), nil
			}
			return nil, nil
		},
	}

	client := ProvideHelmClient(runner)

	err := client.Uninstall("my-release", "my-namespace")
	require.NoError(t, err)
	assert.Equal(t, []string{"uninstall", "my-release", "--namespace", "my-namespace"}, capturedArgs)
}

func TestHelmClient_Uninstall_NoNamespace(t *testing.T) {
	var capturedArgs []string
	runner := &mockCommandRunner{
		runFunc: func(name string, args ...string) ([]byte, error) {
			if name == "helm" && len(args) > 0 && args[0] == "uninstall" {
				capturedArgs = args
				return []byte(""), nil
			}
			return nil, nil
		},
	}

	client := ProvideHelmClient(runner)

	err := client.Uninstall("my-release", "")
	require.NoError(t, err)
	assert.Equal(t, []string{"uninstall", "my-release"}, capturedArgs)
}

func TestHelmClient_List(t *testing.T) {
	var capturedArgs []string
	runner := &mockCommandRunner{
		runFunc: func(name string, args ...string) ([]byte, error) {
			if name == "helm" && len(args) > 0 && args[0] == "list" {
				capturedArgs = args
				return []byte("release1\nrelease2\nrelease3"), nil
			}
			return nil, nil
		},
	}

	client := ProvideHelmClient(runner)

	releases, err := client.List("managed-by=dx", "my-namespace")
	require.NoError(t, err)
	assert.Equal(t, []string{"release1", "release2", "release3"}, releases)
	assert.Equal(t, []string{"list", "-l", "managed-by=dx", "--short", "--namespace", "my-namespace"}, capturedArgs)
}

func TestHelmClient_List_NoNamespace(t *testing.T) {
	var capturedArgs []string
	runner := &mockCommandRunner{
		runFunc: func(name string, args ...string) ([]byte, error) {
			if name == "helm" && len(args) > 0 && args[0] == "list" {
				capturedArgs = args
				return []byte("release1\nrelease2"), nil
			}
			return nil, nil
		},
	}

	client := ProvideHelmClient(runner)

	releases, err := client.List("managed-by=dx", "")
	require.NoError(t, err)
	assert.Equal(t, []string{"release1", "release2"}, releases)
	assert.Equal(t, []string{"list", "-l", "managed-by=dx", "--short"}, capturedArgs)
}

func TestHelmClient_List_Empty(t *testing.T) {
	runner := &mockCommandRunner{
		runFunc: func(name string, args ...string) ([]byte, error) {
			if name == "helm" && len(args) > 0 && args[0] == "list" {
				return []byte(""), nil
			}
			return nil, nil
		},
	}

	client := ProvideHelmClient(runner)

	releases, err := client.List("managed-by=dx", "my-namespace")
	require.NoError(t, err)
	assert.Empty(t, releases)
}

func TestHelmClient_UpgradeFromManifests_Error(t *testing.T) {
	runner := &mockCommandRunner{
		runFunc: func(name string, args ...string) ([]byte, error) {
			return []byte("Error: release failed"), errors.New("exit status 1")
		},
	}

	client := ProvideHelmClient(runner)

	err := client.UpgradeFromManifests("my-release", "my-namespace", "/path/to/wrapper")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "helm upgrade failed")
	assert.Contains(t, err.Error(), "release failed")
}

func TestHelmClient_Uninstall_Error(t *testing.T) {
	runner := &mockCommandRunner{
		runFunc: func(name string, args ...string) ([]byte, error) {
			return []byte("Error: release not found"), errors.New("exit status 1")
		},
	}

	client := ProvideHelmClient(runner)

	err := client.Uninstall("my-release", "my-namespace")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to uninstall helm chart")
	assert.Contains(t, err.Error(), "release not found")
}

func TestHelmClient_List_Error(t *testing.T) {
	runner := &mockCommandRunner{
		runFunc: func(name string, args ...string) ([]byte, error) {
			return []byte("Error: cannot access cluster"), errors.New("exit status 1")
		},
	}

	client := ProvideHelmClient(runner)

	_, err := client.List("managed-by=dx", "my-namespace")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list helm charts")
	assert.Contains(t, err.Error(), "cannot access cluster")
}

func TestHelmClientInterface(t *testing.T) {
	// Verify HelmClient implements the ports.HelmClient interface
	var _ ports.HelmClient = (*HelmClient)(nil)
}
