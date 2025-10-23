package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/example/chat/internal/domain"
	"github.com/example/chat/internal/infrastructure/auth"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrSessionNotFound    = errors.New("session not found")
)

// AuthUseCase defines the interface for authentication use cases
type AuthUseCase interface {
	Register(input RegisterInput) (*AuthOutput, error)
	Login(input LoginInput) (*AuthOutput, error)
	RefreshToken(input RefreshTokenInput) (*AuthOutput, error)
	Logout(input LogoutInput) (*LogoutOutput, error)
}

type authInteractor struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
	jwtService  *auth.JWTService
	passwordSvc *auth.PasswordService

	// Configuration
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewAuthInteractor creates a new auth interactor
func NewAuthInteractor(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	jwtService *auth.JWTService,
	passwordSvc *auth.PasswordService,
) AuthUseCase {
	return &authInteractor{
		userRepo:             userRepo,
		sessionRepo:          sessionRepo,
		jwtService:           jwtService,
		passwordSvc:          passwordSvc,
		accessTokenDuration:  15 * time.Minute,
		refreshTokenDuration: 7 * 24 * time.Hour, // 7 days
	}
}

func (i *authInteractor) Register(input RegisterInput) (*AuthOutput, error) {
	// Check if user already exists
	existing, err := i.userRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := i.passwordSvc.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		Email:        input.Email,
		PasswordHash: hashedPassword,
		DisplayName:  input.DisplayName,
	}

	if err := i.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Generate tokens
	return i.generateAuthOutput(user)
}

func (i *authInteractor) Login(input LoginInput) (*AuthOutput, error) {
	// Find user by email
	user, err := i.userRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := i.passwordSvc.VerifyPassword(input.Password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	return i.generateAuthOutput(user)
}

func (i *authInteractor) RefreshToken(input RefreshTokenInput) (*AuthOutput, error) {
	// Verify refresh token
	claims, err := i.jwtService.VerifyToken(input.RefreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Find user
	user, err := i.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidToken
	}

	// Find active sessions for this user
	sessions, err := i.sessionRepo.FindActiveByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	// Verify that this refresh token exists in active sessions
	validSession := false
	for _, session := range sessions {
		if err := i.passwordSvc.VerifyPassword(input.RefreshToken, session.RefreshTokenHash); err == nil {
			validSession = true
			break
		}
	}

	if !validSession {
		return nil, ErrInvalidToken
	}

	// Generate new tokens
	return i.generateAuthOutput(user)
}

func (i *authInteractor) Logout(input LogoutInput) (*LogoutOutput, error) {
	// Find active sessions
	sessions, err := i.sessionRepo.FindActiveByUserID(input.UserID)
	if err != nil {
		return nil, err
	}

	// Find the session matching this refresh token and revoke it
	for _, session := range sessions {
		if err := i.passwordSvc.VerifyPassword(input.RefreshToken, session.RefreshTokenHash); err == nil {
			if err := i.sessionRepo.Revoke(session.ID); err != nil {
				return nil, err
			}
			return &LogoutOutput{Success: true}, nil
		}
	}

	return nil, ErrSessionNotFound
}

// Helper function to generate auth output with tokens
func (i *authInteractor) generateAuthOutput(user *domain.User) (*AuthOutput, error) {
	// Generate access token
	accessToken, err := i.jwtService.GenerateToken(user.ID, i.accessTokenDuration)
	if err != nil {
		return nil, err
	}

	// Generate refresh token (random secure string)
	refreshToken, err := generateSecureToken()
	if err != nil {
		return nil, err
	}

	// Hash refresh token for storage
	refreshTokenHash, err := i.passwordSvc.HashPassword(refreshToken)
	if err != nil {
		return nil, err
	}

	// Store session
	session := &domain.Session{
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        time.Now().Add(i.refreshTokenDuration),
	}

	if err := i.sessionRepo.Create(session); err != nil {
		return nil, err
	}

	return &AuthOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    session.ExpiresAt,
		User: UserInfo{
			ID:          user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarURL,
		},
	}, nil
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
