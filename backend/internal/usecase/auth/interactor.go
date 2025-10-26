package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainerrors "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

// Service interfaces
type TokenClaims struct {
	UserID string
	Email  string
}

type JWTService interface {
	GenerateToken(userID string, duration time.Duration) (string, error)
	VerifyToken(token string) (*TokenClaims, error)
}

type PasswordService interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) error
}

var (
	ErrInvalidCredentials = domainerrors.ErrInvalidCredentials
	ErrUserAlreadyExists  = domainerrors.ErrUserAlreadyExists
	ErrInvalidToken       = domainerrors.ErrInvalidToken
	ErrSessionNotFound    = domainerrors.ErrSessionNotFound
)

// AuthUseCase defines the interface for authentication use cases
type AuthUseCase interface {
	Register(ctx context.Context, input RegisterInput) (*AuthOutput, error)
	Login(ctx context.Context, input LoginInput) (*AuthOutput, error)
	RefreshToken(ctx context.Context, input RefreshTokenInput) (*AuthOutput, error)
	Logout(ctx context.Context, input LogoutInput) (*LogoutOutput, error)
}

type authInteractor struct {
	userRepo    domainrepository.UserRepository
	sessionRepo domainrepository.SessionRepository
	jwtService  JWTService
	passwordSvc PasswordService

	// Configuration
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewAuthInteractor creates a new auth interactor
func NewAuthInteractor(
	userRepo domainrepository.UserRepository,
	sessionRepo domainrepository.SessionRepository,
	jwtService JWTService,
	passwordSvc PasswordService,
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

func (i *authInteractor) Register(ctx context.Context, input RegisterInput) (*AuthOutput, error) {
	// Check if user already exists
	existing, err := i.userRepo.FindByEmail(ctx, input.Email)
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
	user := &entity.User{
		Email:        input.Email,
		PasswordHash: hashedPassword,
		DisplayName:  input.DisplayName,
	}

	if err := i.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate tokens
	return i.generateAuthOutput(ctx, user)
}

func (i *authInteractor) Login(ctx context.Context, input LoginInput) (*AuthOutput, error) {
	// Find user by email
	user, err := i.userRepo.FindByEmail(ctx, input.Email)
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
	return i.generateAuthOutput(ctx, user)
}

func (i *authInteractor) RefreshToken(ctx context.Context, input RefreshTokenInput) (*AuthOutput, error) {
	// Verify refresh token
	claims, err := i.jwtService.VerifyToken(input.RefreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Find user
	user, err := i.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidToken
	}

	// Find active sessions for this user
	sessions, err := i.sessionRepo.FindActiveByUserID(ctx, user.ID)
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
	return i.generateAuthOutput(ctx, user)
}

func (i *authInteractor) Logout(ctx context.Context, input LogoutInput) (*LogoutOutput, error) {
	// Find active sessions
	sessions, err := i.sessionRepo.FindActiveByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	// Find the session matching this refresh token and revoke it
	for _, session := range sessions {
		if err := i.passwordSvc.VerifyPassword(input.RefreshToken, session.RefreshTokenHash); err == nil {
			if err := i.sessionRepo.Revoke(ctx, session.ID); err != nil {
				return nil, err
			}
			return &LogoutOutput{Success: true}, nil
		}
	}

	return nil, ErrSessionNotFound
}

// Helper function to generate auth output with tokens
func (i *authInteractor) generateAuthOutput(ctx context.Context, user *entity.User) (*AuthOutput, error) {
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
	session := &entity.Session{
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        time.Now().Add(i.refreshTokenDuration),
	}

	if err := i.sessionRepo.Create(ctx, session); err != nil {
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
