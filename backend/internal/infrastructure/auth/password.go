package auth

import (
	authuc "github.com/newt239/chat/internal/usecase/auth"
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
