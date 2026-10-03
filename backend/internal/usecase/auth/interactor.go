package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domaintransaction "github.com/newt239/chat/internal/domain/transaction"
	"github.com/newt239/chat/internal/usecase/audit"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 30 * 24 * time.Hour
)

// TokenClaims はアクセストークンに載せる本人とセッションです
type TokenClaims struct {
	UserID    string
	SessionID string
}

type JWTService interface {
	GenerateToken(claims TokenClaims, duration time.Duration) (string, error)
	VerifyToken(token string) (*TokenClaims, error)
}

// SessionCloser は失効したセッションのリアルタイム接続を切ります
type SessionCloser interface {
	CloseSession(sessionID string)
}

type PasswordService interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) error
}

// GoogleVerifier は ID トークンの署名・audience・有効期限を検証します
type GoogleVerifier interface {
	Verify(ctx context.Context, idToken string) (*GoogleIdentity, error)
}

// GoogleCodeExchanger は認可コードと PKCE の code_verifier を Google の ID トークンと交換します
type GoogleCodeExchanger interface {
	Exchange(ctx context.Context, code, codeVerifier string) (idToken string, err error)
}

type Interactor struct {
	userRepo            domainrepository.UserRepository
	sessionRepo         domainrepository.SessionRepository
	workspaceRepo       domainrepository.WorkspaceRepository
	invitationRepo      domainrepository.InvitationRepository
	jwtService          JWTService
	passwordSvc         PasswordService
	googleVerifier      GoogleVerifier
	googleCode          GoogleCodeExchanger
	txManager           domaintransaction.Manager
	recorder            audit.Recorder
	sessionCloser       SessionCloser
	passwordAuthEnabled bool
}

func New(
	userRepo domainrepository.UserRepository,
	sessionRepo domainrepository.SessionRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	invitationRepo domainrepository.InvitationRepository,
	jwtService JWTService,
	passwordSvc PasswordService,
	googleVerifier GoogleVerifier,
	googleCode GoogleCodeExchanger,
	txManager domaintransaction.Manager,
	recorder audit.Recorder,
	sessionCloser SessionCloser,
	passwordAuthEnabled bool,
) *Interactor {
	return &Interactor{
		userRepo:            userRepo,
		sessionRepo:         sessionRepo,
		workspaceRepo:       workspaceRepo,
		invitationRepo:      invitationRepo,
		jwtService:          jwtService,
		passwordSvc:         passwordSvc,
		googleVerifier:      googleVerifier,
		googleCode:          googleCode,
		txManager:           txManager,
		recorder:            recorder,
		sessionCloser:       sessionCloser,
		passwordAuthEnabled: passwordAuthEnabled,
	}
}

func (i *Interactor) PasswordAuthEnabled() bool {
	return i.passwordAuthEnabled
}

func (i *Interactor) Login(ctx context.Context, input LoginInput) (*AuthOutput, error) {
	if !i.passwordAuthEnabled {
		return nil, domerr.ErrPasswordAuthDisabled
	}
	user, err := i.userRepo.FindByEmail(ctx, entity.NormalizeEmail(input.Email))
	if err != nil {
		return nil, err
	}
	if user == nil || user.IsApp {
		return nil, domerr.ErrInvalidCredentials
	}

	if err := i.passwordSvc.VerifyPassword(input.Password, user.PasswordHash); err != nil {
		i.recordLogin(ctx, user, entity.AuditActionLoginFailed)
		return nil, domerr.ErrInvalidCredentials
	}
	return i.login(ctx, user)
}

// LoginWithGoogle は sub で照合し、未紐付けならメールアドレスで既存ユーザーに紐付け、どちらもなければ招待か登録を許可したワークスペースがある場合だけユーザーを作ります
func (i *Interactor) LoginWithGoogle(ctx context.Context, input LoginWithGoogleInput) (*AuthOutput, error) {
	identity, err := i.googleVerifier.Verify(ctx, input.IDToken)
	if err != nil {
		return nil, err
	}
	return i.loginWithGoogleIdentity(ctx, identity, input.WorkspaceID)
}

// LoginWithGoogleCode はネイティブアプリがブラウザで受け取った認可コードを交換し、LoginWithGoogle と同じ扱いでログインさせます
func (i *Interactor) LoginWithGoogleCode(ctx context.Context, input LoginWithGoogleCodeInput) (*AuthOutput, error) {
	idToken, err := i.googleCode.Exchange(ctx, input.Code, input.CodeVerifier)
	if err != nil {
		return nil, err
	}
	identity, err := i.googleVerifier.Verify(ctx, idToken)
	if err != nil {
		return nil, err
	}
	// 別の認可リクエストで発行されたトークンを使い回させない
	if identity.Nonce == "" || identity.Nonce != input.Nonce {
		return nil, domerr.ErrInvalidToken
	}
	return i.loginWithGoogleIdentity(ctx, identity, input.WorkspaceID)
}

func (i *Interactor) loginWithGoogleIdentity(ctx context.Context, identity *GoogleIdentity, workspaceID *string) (*AuthOutput, error) {
	if !identity.EmailVerified {
		return nil, domerr.ErrEmailNotVerified
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
		user, err = i.createGoogleUser(ctx, identity, workspaceID)
	} else if workspaceID != nil {
		err = i.joinSignupWorkspace(ctx, user.ID, *workspaceID)
	}
	if err != nil {
		return nil, err
	}
	return i.login(ctx, user)
}

func (i *Interactor) linkGoogleAccount(ctx context.Context, identity *GoogleIdentity) (*entity.User, error) {
	user, err := i.userRepo.FindByEmail(ctx, entity.NormalizeEmail(identity.Email))
	if err != nil || user == nil {
		return nil, err
	}
	// 別の Google アカウントに紐付いたユーザーやボットは乗っ取れないようにする
	if user.IsApp || user.GoogleSub != nil {
		return nil, domerr.ErrInvalidCredentials
	}
	user.GoogleSub = &identity.Sub
	if err := i.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (i *Interactor) createGoogleUser(ctx context.Context, identity *GoogleIdentity, workspaceID *string) (*entity.User, error) {
	email := entity.NormalizeEmail(identity.Email)
	invitations, err := i.invitationRepo.FindPendingByEmail(ctx, email, time.Now())
	if err != nil {
		return nil, err
	}
	if len(invitations) == 0 && workspaceID == nil {
		return nil, domerr.ErrInvitationRequired
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

func (i *Interactor) SignUp(ctx context.Context, input SignUpInput) (*AuthOutput, error) {
	if !i.passwordAuthEnabled {
		return nil, domerr.ErrPasswordAuthDisabled
	}
	email := entity.NormalizeEmail(input.Email)
	existing, err := i.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domerr.ErrUserAlreadyExists
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
func (i *Interactor) createSignupUser(ctx context.Context, user *entity.User, workspaceID string, byEmail bool) error {
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
func (i *Interactor) joinSignupWorkspace(ctx context.Context, userID, workspaceID string) error {
	if err := i.checkSignupEnabled(ctx, workspaceID, false); err != nil {
		return err
	}
	// 停止中のメンバーは参加し直させない
	member, err := i.workspaceRepo.FindMemberIncludingSuspended(ctx, workspaceID, userID)
	if err != nil || member != nil {
		return err
	}
	return i.addSignupMember(ctx, workspaceID, userID)
}

func (i *Interactor) checkSignupEnabled(ctx context.Context, workspaceID string, byEmail bool) error {
	ws, err := i.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return err
	}
	if ws == nil || !ws.SignupEnabled || (byEmail && !ws.EmailSignupEnabled) {
		return domerr.ErrSignupDisabled
	}
	return nil
}

func (i *Interactor) addSignupMember(ctx context.Context, workspaceID, userID string) error {
	return i.workspaceRepo.AddMember(ctx, &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: userID, Role: entity.WorkspaceRoleMember})
}

func (i *Interactor) SignUpWithInvitation(ctx context.Context, input SignUpWithInvitationInput) (*AuthOutput, error) {
	if !i.passwordAuthEnabled {
		return nil, domerr.ErrPasswordAuthDisabled
	}
	now := time.Now()
	invitation, err := i.invitationRepo.FindByTokenHash(ctx, entity.HashSecretToken(input.Token))
	if err != nil {
		return nil, err
	}
	if invitation == nil || !invitation.IsPending(now) {
		return nil, domerr.ErrInvitationNotFound
	}
	existing, err := i.userRepo.FindByEmail(ctx, invitation.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domerr.ErrUserAlreadyExists
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
func (i *Interactor) createInvitedUser(ctx context.Context, user *entity.User, invitations []*entity.Invitation) error {
	return i.txManager.Do(ctx, func(ctx context.Context) error {
		if err := i.userRepo.Create(ctx, user); err != nil {
			return err
		}
		now := time.Now()
		joined := make(map[string]bool, len(invitations))
		// 同じワークスペースへの招待が複数あれば、新しい順に並んでいるため最新のロールで参加する
		for _, inv := range invitations {
			if !joined[inv.WorkspaceID] {
				member := &entity.WorkspaceMember{WorkspaceID: inv.WorkspaceID, UserID: user.ID, Role: inv.Role}
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

func (i *Interactor) login(ctx context.Context, user *entity.User) (*AuthOutput, error) {
	out, err := i.createSession(ctx, user)
	if err != nil {
		return nil, err
	}
	i.recordLogin(ctx, user, entity.AuditActionLogin)
	return out, nil
}

// recordLogin はログインがワークスペースに属さないため、ユーザーが参加している全ワークスペースの監査ログに記録します
func (i *Interactor) recordLogin(ctx context.Context, user *entity.User, action entity.AuditAction) {
	memberships, err := i.workspaceRepo.FindMembershipsByUserID(ctx, user.ID)
	if err != nil {
		return
	}
	// 失敗したログインは本人の操作とは限らないため実行者を空にする
	var actorID *string
	if action == entity.AuditActionLogin {
		actorID = &user.ID
	}
	for _, m := range memberships {
		i.recorder.Record(ctx, entity.AuditLog{
			WorkspaceID: m.WorkspaceID,
			ActorID:     actorID,
			Action:      action,
			TargetType:  entity.AuditTargetUser,
			TargetID:    user.ID,
			TargetLabel: user.Email,
		})
	}
}

func (i *Interactor) RefreshToken(ctx context.Context, input RefreshTokenInput) (*AuthOutput, error) {
	if input.RefreshToken == "" {
		return nil, domerr.ErrInvalidToken
	}
	session, err := i.sessionRepo.FindActiveByTokenHash(ctx, entity.HashSecretToken(input.RefreshToken))
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, domerr.ErrInvalidToken
	}
	user, err := i.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domerr.ErrInvalidToken
	}

	// セッションはログイン単位で保持し、リフレッシュではトークンだけを差し替える
	tokens, err := i.issueTokens(user, session.ID)
	if err != nil {
		return nil, err
	}
	if err := i.sessionRepo.Rotate(ctx, session.ID, tokens.refreshTokenHash, tokens.output.ExpiresAt); err != nil {
		return nil, err
	}
	return tokens.output, nil
}

// Logout はリフレッシュトークンかアクセストークンが指すセッションだけを失効させ、その接続を切ります
func (i *Interactor) Logout(ctx context.Context, input LogoutInput) error {
	sessionID := input.SessionID
	if input.RefreshToken != "" {
		session, err := i.sessionRepo.FindActiveByTokenHash(ctx, entity.HashSecretToken(input.RefreshToken))
		if err != nil {
			return err
		}
		if session != nil {
			sessionID = session.ID
		}
	}
	if sessionID == "" {
		return nil
	}
	if err := i.sessionRepo.Revoke(ctx, sessionID); err != nil {
		return err
	}
	i.sessionCloser.CloseSession(sessionID)
	return nil
}

type issuedTokens struct {
	output           *AuthOutput
	refreshTokenHash string
}

func (i *Interactor) issueTokens(user *entity.User, sessionID string) (*issuedTokens, error) {
	accessToken, err := i.jwtService.GenerateToken(TokenClaims{UserID: user.ID, SessionID: sessionID}, accessTokenTTL)
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
			ExpiresAt:    time.Now().Add(refreshTokenTTL),
			User:         user,
		},
		refreshTokenHash: refreshTokenHash,
	}, nil
}

// createSession はアクセストークンに載せるためセッション ID を先に採番し、ログイン元の端末情報とともに保存します
func (i *Interactor) createSession(ctx context.Context, user *entity.User) (*AuthOutput, error) {
	sessionID := uuid.NewString()
	tokens, err := i.issueTokens(user, sessionID)
	if err != nil {
		return nil, err
	}

	client := audit.ClientInfoFrom(ctx)
	session := &entity.Session{
		ID:               sessionID,
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
