package auth

import (
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// PasswordService provides password hashing and verification
type PasswordService struct {
	cost int
}

// NewPasswordService creates a new password service
func NewPasswordService() *PasswordService {
	return &PasswordService{
		cost: bcryptCost,
	}
}

// HashPassword hashes a password using bcrypt
func (s *PasswordService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against a hash
func (s *PasswordService) VerifyPassword(password, hashedPassword string) error {
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
