package testutil

import (
	"github.com/stretchr/testify/mock"
)

// MockSymmetricEncryptor provides a testify mock for ports.SymmetricEncryptor
type MockSymmetricEncryptor struct {
	mock.Mock
}

func (m *MockSymmetricEncryptor) Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	args := m.Called(plaintext, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockSymmetricEncryptor) Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	args := m.Called(ciphertext, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockSymmetricEncryptor) CreateKey() ([]byte, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}
