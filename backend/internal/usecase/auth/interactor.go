package auth

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainerrors "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domaintransaction "github.com/newt239/chat/internal/domain/transaction"
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

// GoogleVerifier は ID トークンの署名・audience・有効期限を検証します
type GoogleVerifier interface {
	Verify(ctx context.Context, idToken string) (*GoogleIdentity, error)
}

var (
	ErrInvalidCredentials   = domainerrors.ErrInvalidCredentials
	ErrUserAlreadyExists    = domainerrors.ErrUserAlreadyExists
	ErrInvalidToken         = domainerrors.ErrInvalidToken
	ErrSessionNotFound      = domainerrors.ErrSessionNotFound
	ErrInvitationRequired   = domainerrors.ErrInvitationRequired
	ErrInvitationNotFound   = domainerrors.ErrInvitationNotFound
	ErrEmailNotVerified     = domainerrors.ErrEmailNotVerified
	ErrPasswordAuthDisabled = domainerrors.ErrPasswordAuthDisabled
	ErrSignupDisabled       = domainerrors.ErrSignupDisabled
)

type AuthUseCase interface {
	PasswordAuthEnabled() bool
	Login(ctx context.Context, input LoginInput) (*AuthOutput, error)
	LoginWithGoogle(ctx context.Context, input LoginWithGoogleInput) (*AuthOutput, error)
	SignUp(ctx context.Context, input SignUpInput) (*AuthOutput, error)
	SignUpWithInvitation(ctx context.Context, input SignUpWithInvitationInput) (*AuthOutput, error)
	RefreshToken(ctx context.Context, input RefreshTokenInput) (*AuthOutput, error)
	Logout(ctx context.Context, input LogoutInput) (*LogoutOutput, error)
}

type authInteractor struct {
	userRepo       domainrepository.UserRepository
	sessionRepo    domainrepository.SessionRepository
	workspaceRepo  domainrepository.WorkspaceRepository
	invitationRepo domainrepository.InvitationRepository
	jwtService     JWTService
	passwordSvc    PasswordService
	googleVerifier GoogleVerifier
	txManager      domaintransaction.Manager
	recorder       audit.Recorder
	settings       Settings
}

func NewAuthInteractor(
	userRepo domainrepository.UserRepository,
	sessionRepo domainrepository.SessionRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	invitationRepo domainrepository.InvitationRepository,
	jwtService JWTService,
	passwordSvc PasswordService,
	googleVerifier GoogleVerifier,
	txManager domaintransaction.Manager,
	recorder audit.Recorder,
	settings Settings,
) AuthUseCase {
	return &authInteractor{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		workspaceRepo:  workspaceRepo,
		invitationRepo: invitationRepo,
		jwtService:     jwtService,
		passwordSvc:    passwordSvc,
		googleVerifier: googleVerifier,
		txManager:      txManager,
		recorder:       recorder,
		settings:       settings,
	}
}

func (i *authInteractor) PasswordAuthEnabled() bool {
	return i.settings.PasswordAuthEnabled
}

func (i *authInteractor) Login(ctx context.Context, input LoginInput) (*AuthOutput, error) {
	if !i.settings.PasswordAuthEnabled {
		return nil, ErrPasswordAuthDisabled
	}
	user, err := i.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if user == nil || user.IsBot {
		return nil, ErrInvalidCredentials
	}

	if err := i.passwordSvc.VerifyPassword(input.Password, user.PasswordHash); err != nil {
		i.recordLogin(ctx, user, entity.AuditActionLoginFailed)
		return nil, ErrInvalidCredentials
	}
	return i.login(ctx, user)
}

// LoginWithGoogle は sub で照合し、未紐付けならメールアドレスで既存ユーザーに紐付け、どちらもなければ招待か登録を許可したワークスペースがある場合だけユーザーを作ります
func (i *authInteractor) LoginWithGoogle(ctx context.Context, input LoginWithGoogleInput) (*AuthOutput, error) {
	identity, err := i.googleVerifier.Verify(ctx, input.IDToken)
	if err != nil {
		return nil, err
	}
	if !identity.EmailVerified {
		return nil, ErrEmailNotVerified
	}

	user, err := i.userRepo.FindByGoogleSub(ctx, identity.Sub)
	if err != nil {
		return nil, err
	}
	if user == nil {
		user, err = i.linkGoogleAccount(ctx, identity)
		if err != nil {
			return nil, err
		}
	}
	if user == nil {
		user, err = i.createGoogleUser(ctx, identity, input.WorkspaceID)
	} else if input.WorkspaceID != nil {
		err = i.joinSignupWorkspace(ctx, user.ID, *input.WorkspaceID)
	}
	if err != nil {
		return nil, err
	}
	return i.login(ctx, user)
}

func (i *authInteractor) linkGoogleAccount(ctx context.Context, identity *GoogleIdentity) (*entity.User, error) {
	user, err := i.userRepo.FindByEmail(ctx, entity.NormalizeEmail(identity.Email))
	if err != nil || user == nil {
		return nil, err
	}
	// 別の Google アカウントに紐付いたユーザーやボットは乗っ取れないようにする
	if user.IsBot || user.GoogleSub != nil {
		return nil, ErrInvalidCredentials
	}
	user.GoogleSub = &identity.Sub
	if err := i.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (i *authInteractor) createGoogleUser(ctx context.Context, identity *GoogleIdentity, workspaceID *string) (*entity.User, error) {
	email := entity.NormalizeEmail(identity.Email)
	invitations, err := i.invitationRepo.FindPendingByEmail(ctx, email, time.Now())
	if err != nil {
		return nil, err
	}
	if len(invitations) == 0 && workspaceID == nil {
		return nil, ErrInvitationRequired
	}

	user := &entity.User{
		Email:        email,
		PasswordHash: entity.UnusablePasswordHash,
		GoogleSub:    &identity.Sub,
		DisplayName:  identity.Name,
	}
	if user.DisplayName == "" {
		user.DisplayName = email
	}
	if identity.Picture != "" {
		user.AvatarURL = &identity.Picture
	}
	if len(invitations) > 0 {
		err = i.createInvitedUser(ctx, user, invitations)
	} else {
		err = i.createSignupUser(ctx, user, *workspaceID, false)
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (i *authInteractor) SignUp(ctx context.Context, input SignUpInput) (*AuthOutput, error) {
	if !i.settings.PasswordAuthEnabled {
		return nil, ErrPasswordAuthDisabled
	}
	email := entity.NormalizeEmail(input.Email)
	existing, err := i.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}
	hashed, err := i.passwordSvc.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}
	user := &entity.User{Email: email, PasswordHash: hashed, DisplayName: input.DisplayName}
	if err := i.createSignupUser(ctx, user, input.WorkspaceID, true); err != nil {
		return nil, err
	}
	return i.login(ctx, user)
}

// createSignupUser は登録を許可したワークスペースにメンバーとして参加させる形でユーザーを作ります
func (i *authInteractor) createSignupUser(ctx context.Context, user *entity.User, workspaceID string, byEmail bool) error {
	if err := i.checkSignupEnabled(ctx, workspaceID, byEmail); err != nil {
		return err
	}
	return i.txManager.Do(ctx, func(ctx context.Context) error {
		if err := i.userRepo.Create(ctx, user); err != nil {
			return err
		}
		return i.addSignupMember(ctx, workspaceID, user.ID)
	})
}

// joinSignupWorkspace は参加リンクから既存のアカウントでログインしたとき、未参加ならメンバーとして参加させます
func (i *authInteractor) joinSignupWorkspace(ctx context.Context, userID, workspaceID string) error {
	if err := i.checkSignupEnabled(ctx, workspaceID, false); err != nil {
		return err
	}
	member, err := i.workspaceRepo.FindMember(ctx, workspaceID, userID)
	if err != nil || member != nil {
		return err
	}
	return i.addSignupMember(ctx, workspaceID, userID)
}

func (i *authInteractor) checkSignupEnabled(ctx context.Context, workspaceID string, byEmail bool) error {
	ws, err := i.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return err
	}
	if ws == nil || !ws.SignupEnabled || (byEmail && !ws.EmailSignupEnabled) {
		return ErrSignupDisabled
	}
	return nil
}

func (i *authInteractor) addSignupMember(ctx context.Context, workspaceID, userID string) error {
	return i.workspaceRepo.AddMember(ctx, &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: userID, Role: entity.WorkspaceRoleMember, JoinedAt: time.Now()})
}

func (i *authInteractor) SignUpWithInvitation(ctx context.Context, input SignUpWithInvitationInput) (*AuthOutput, error) {
	if !i.settings.PasswordAuthEnabled {
		return nil, ErrPasswordAuthDisabled
	}
	now := time.Now()
	invitation, err := i.invitationRepo.FindByTokenHash(ctx, entity.HashSecretToken(input.Token))
	if err != nil {
		return nil, err
	}
	if invitation == nil || !invitation.IsPending(now) {
		return nil, ErrInvitationNotFound
	}
	existing, err := i.userRepo.FindByEmail(ctx, invitation.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	hashed, err := i.passwordSvc.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}
	// 同じメールアドレスへの他のワークスペースからの招待もまとめて受諾する
	invitations, err := i.invitationRepo.FindPendingByEmail(ctx, invitation.Email, now)
	if err != nil {
		return nil, err
	}
	user := &entity.User{Email: invitation.Email, PasswordHash: hashed, DisplayName: input.DisplayName}
	if err := i.createInvitedUser(ctx, user, invitations); err != nil {
		return nil, err
	}
	return i.login(ctx, user)
}

// createInvitedUser はユーザーを作り、招待されたワークスペースに参加させます
func (i *authInteractor) createInvitedUser(ctx context.Context, user *entity.User, invitations []*entity.Invitation) error {
	return i.txManager.Do(ctx, func(ctx context.Context) error {
		if err := i.userRepo.Create(ctx, user); err != nil {
			return err
		}
		now := time.Now()
		joined := make(map[string]bool, len(invitations))
		// 同じワークスペースへの招待が複数あれば、新しい順に並んでいるため最新のロールで参加する
		for _, inv := range invitations {
			if !joined[inv.WorkspaceID] {
				member := &entity.WorkspaceMember{WorkspaceID: inv.WorkspaceID, UserID: user.ID, Role: inv.Role, JoinedAt: now}
				if err := i.workspaceRepo.AddMember(ctx, member); err != nil {
					return err
				}
				joined[inv.WorkspaceID] = true
			}
			if err := i.invitationRepo.MarkAccepted(ctx, inv.ID, now); err != nil {
				return err
			}
		}
		return nil
	})
}

func (i *authInteractor) login(ctx context.Context, user *entity.User) (*AuthOutput, error) {
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
	session, err := i.sessionRepo.FindActiveByTokenHash(ctx, entity.HashSecretToken(input.RefreshToken))
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrInvalidToken
	}
	user, err := i.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidToken
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
	accessToken, err := i.jwtService.GenerateToken(user.ID, i.settings.AccessTokenTTL)
	if err != nil {
		return nil, err
	}

	// リフレッシュトークンは JWT ではなく乱数で、セッションを引けるよう SHA-256 で保存する
	refreshToken, refreshTokenHash, err := entity.NewSecretToken()
	if err != nil {
		return nil, err
	}

	return &issuedTokens{
		output: &AuthOutput{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(i.settings.RefreshTokenTTL),
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
