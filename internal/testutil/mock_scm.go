package testutil

import (
	"github.com/stretchr/testify/mock"
)

type MockScm struct {
	mock.Mock
}

func (m *MockScm) Download(repositoryUrl string, branch string, repositoryPath string) error {
	args := m.Called(repositoryUrl, branch, repositoryPath)
	return args.Error(0)
}
