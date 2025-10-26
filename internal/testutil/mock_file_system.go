package testutil

import (
	"dx/internal/ports"
	"github.com/stretchr/testify/mock"
)

type MockFileSystem struct {
	mock.Mock
}

func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	args := m.Called(path)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockFileSystem) WriteFile(path string, content []byte, accessMode ports.AccessMode) error {
	args := m.Called(path, content, accessMode)
	return args.Error(0)
}

func (m *MockFileSystem) EnsureDirExists(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockFileSystem) FileExists(path string) (bool, error) {
	args := m.Called(path)
	return args.Bool(0), args.Error(1)
}
