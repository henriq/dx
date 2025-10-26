package testutil

import (
	"dx/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

type MockContainerImageRepository struct {
	mock.Mock
}

func (m *MockContainerImageRepository) BuildImage(image domain.DockerImage) error {
	args := m.Called(image)
	return args.Error(0)
}

func (m *MockContainerImageRepository) PullImage(image string) error {
	args := m.Called(image)
	return args.Error(0)
}
