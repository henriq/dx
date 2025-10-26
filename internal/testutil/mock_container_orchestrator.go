package testutil

import (
	"dx/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

type MockContainerOrchestrator struct {
	mock.Mock
}

func (m *MockContainerOrchestrator) CreateClusterEnvironmentKey() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockContainerOrchestrator) InstallService(service *domain.Service) error {
	args := m.Called(service)
	return args.Error(0)
}

func (m *MockContainerOrchestrator) InstallDevProxy(service *domain.Service) error {
	args := m.Called(service)
	return args.Error(0)
}

func (m *MockContainerOrchestrator) UninstallService(service *domain.Service) error {
	args := m.Called(service)
	return args.Error(0)
}

func (m *MockContainerOrchestrator) HasDeployedServices() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}
