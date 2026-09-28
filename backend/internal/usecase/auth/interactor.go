package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainerrors "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/usecase/audit"
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
	userRepo      domainrepository.UserRepository
	sessionRepo   domainrepository.SessionRepository
	workspaceRepo domainrepository.WorkspaceRepository
	jwtService    JWTService
	passwordSvc   PasswordService
	recorder      audit.Recorder

	// Configuration
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewAuthInteractor creates a new auth interactor
func NewAuthInteractor(
	userRepo domainrepository.UserRepository,
	sessionRepo domainrepository.SessionRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	jwtService JWTService,
	passwordSvc PasswordService,
	recorder audit.Recorder,
) AuthUseCase {
	return &authInteractor{
		userRepo:             userRepo,
		sessionRepo:          sessionRepo,
		workspaceRepo:        workspaceRepo,
		jwtService:           jwtService,
		passwordSvc:          passwordSvc,
		recorder:             recorder,
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

	return i.createSession(ctx, user)
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
		i.recordLogin(ctx, user, entity.AuditActionLoginFailed)
		return nil, ErrInvalidCredentials
	}

	out, err := i.createSession(ctx, user)
	if err != nil {
		return nil, err
	}
	i.recordLogin(ctx, user, entity.AuditActionLogin)
	return out, nil
}

// recordLogin はログインがワークスペースに属さないため、ユーザーが参加している全ワークスペースの監査ログに記録します
func (i *authInteractor) recordLogin(ctx context.Context, user *entity.User, action entity.AuditAction) {
	workspaces, err := i.workspaceRepo.FindByUserID(ctx, user.ID)
	if err != nil {
		return
	}
	// 失敗したログインは本人の操作とは限らないため実行者を空にする
	var actorID *string
	if action == entity.AuditActionLogin {
		actorID = &user.ID
	}
	for _, ws := range workspaces {
		i.recorder.Record(ctx, entity.AuditLog{
			WorkspaceID: ws.ID,
			ActorID:     actorID,
			Action:      action,
			TargetType:  entity.AuditTargetUser,
			TargetID:    user.ID,
			TargetLabel: user.Email,
		})
	}
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

	for _, session := range sessions {
		if err := i.passwordSvc.VerifyPassword(input.RefreshToken, session.RefreshTokenHash); err != nil {
			continue
		}
		// セッションはログイン単位で保持し、リフレッシュではトークンだけを差し替える
		tokens, err := i.issueTokens(user)
		if err != nil {
			return nil, err
		}
		if err := i.sessionRepo.Rotate(ctx, session.ID, tokens.refreshTokenHash, tokens.output.ExpiresAt); err != nil {
			return nil, err
		}
		return tokens.output, nil
	}

	return nil, ErrInvalidToken
}

func (i *authInteractor) Logout(ctx context.Context, input LogoutInput) (*LogoutOutput, error) {
	if err := i.sessionRepo.RevokeAllByUserID(ctx, input.UserID); err != nil {
		return nil, err
	}

	return &LogoutOutput{Success: true}, nil
}

type issuedTokens struct {
	output           *AuthOutput
	refreshTokenHash string
}

func (i *authInteractor) issueTokens(user *entity.User) (*issuedTokens, error) {
	accessToken, err := i.jwtService.GenerateToken(user.ID, i.accessTokenDuration)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateSecureToken()
	if err != nil {
		return nil, err
	}

	refreshTokenHash, err := i.passwordSvc.HashPassword(refreshToken)
	if err != nil {
		return nil, err
	}

	return &issuedTokens{
		output: &AuthOutput{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(i.refreshTokenDuration),
			User: UserInfo{
				ID:          user.ID,
				Email:       user.Email,
				DisplayName: user.DisplayName,
				AvatarURL:   user.AvatarURL,
			},
		},
		refreshTokenHash: refreshTokenHash,
	}, nil
}

// createSession はトークンを発行し、ログイン元の端末情報とともにセッションを保存します
func (i *authInteractor) createSession(ctx context.Context, user *entity.User) (*AuthOutput, error) {
	tokens, err := i.issueTokens(user)
	if err != nil {
		return nil, err
	}

	client := audit.ClientInfoFrom(ctx)
	session := &entity.Session{
		UserID:           user.ID,
		RefreshTokenHash: tokens.refreshTokenHash,
		ExpiresAt:        tokens.output.ExpiresAt,
		IPAddress:        client.IPAddress,
		UserAgent:        client.UserAgent,
	}
	if err := i.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}
	return tokens.output, nil
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
