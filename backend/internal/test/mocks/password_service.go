package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockPasswordService はテスト用のPasswordServiceモックです
type MockPasswordService struct {
	mock.Mock
}

func (m *MockPasswordService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordService) VerifyPassword(password, hash string) error {
	args := m.Called(password, hash)
	return args.Error(0)
}
