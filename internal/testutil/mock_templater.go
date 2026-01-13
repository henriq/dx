package testutil

import (
	"github.com/stretchr/testify/mock"
)

type MockTemplater struct {
	mock.Mock
}

func (m *MockTemplater) Render(template string, templateName string, values map[string]interface{}) (string, error) {
	args := m.Called(template, templateName, values)
	return args.String(0), args.Error(1)
}
