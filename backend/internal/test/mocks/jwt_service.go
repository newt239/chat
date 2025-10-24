package mocks

import (
	"time"

	"github.com/example/chat/internal/usecase/auth"
	"github.com/stretchr/testify/mock"
)

// MockJWTService はテスト用のJWTServiceモックです
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateToken(userID string, duration time.Duration) (string, error) {
	args := m.Called(userID, duration)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) VerifyToken(token string) (*auth.TokenClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenClaims), args.Error(1)
}

// TestTokenClaims はテスト用のトークンクレームを作成するヘルパー関数です
func TestTokenClaims(userID, email string) *auth.TokenClaims {
	return &auth.TokenClaims{
		UserID: userID,
		Email:  email,
	}
}
