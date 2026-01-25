package testutil

import (
	"io"

	"github.com/stretchr/testify/mock"
)

// MockCommandRunner provides a testify mock for ports.CommandRunner
type MockCommandRunner struct {
	mock.Mock
}

func (m *MockCommandRunner) Run(name string, args ...string) ([]byte, error) {
	callArgs := m.Called(name, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).([]byte), callArgs.Error(1)
}

func (m *MockCommandRunner) RunWithEnv(name string, env []string, args ...string) ([]byte, error) {
	callArgs := m.Called(name, env, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).([]byte), callArgs.Error(1)
}

func (m *MockCommandRunner) RunInDir(dir, name string, args ...string) ([]byte, error) {
	callArgs := m.Called(dir, name, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).([]byte), callArgs.Error(1)
}

func (m *MockCommandRunner) RunWithEnvInDir(dir string, env []string, name string, args ...string) ([]byte, error) {
	callArgs := m.Called(dir, env, name, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).([]byte), callArgs.Error(1)
}

func (m *MockCommandRunner) RunWithStdin(stdin io.Reader, name string, args ...string) ([]byte, error) {
	callArgs := m.Called(stdin, name, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).([]byte), callArgs.Error(1)
}

func (m *MockCommandRunner) RunInteractive(name string, args ...string) error {
	callArgs := m.Called(name, args)
	return callArgs.Error(0)
}
