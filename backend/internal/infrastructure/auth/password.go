package auth

import (
	authuc "github.com/example/chat/internal/usecase/auth"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

type passwordService struct {
	cost int
}

func NewPasswordService() authuc.PasswordService {
	return &passwordService{
		cost: bcryptCost,
	}
}

func (s *passwordService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *passwordService) VerifyPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Legacy functions for backward compatibility
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
