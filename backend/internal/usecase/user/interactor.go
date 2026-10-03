package user

import (
	"context"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/usecase/auth"
)

type UpdateMeInput struct {
	UserID      string
	DisplayName *string
	Bio         *string
	AvatarURL   *string
	// nil なら変えない。空ならすべて外す
	Links *[]string
}

// UserCloser は全セッションを失効させたユーザーのリアルタイム接続を切ります
type UserCloser interface {
	CloseUser(userID string)
}

// ErrOwnerCannotDelete はオーナーのいないワークスペースを残さないためのエラーです
var ErrOwnerCannotDelete = domerr.New(domerr.ErrFailedPrecondition, "ワークスペースのオーナーはアカウントを削除できません。先にワークスペースを削除してください")

type Interactor struct {
	userRepo      domainrepository.UserRepository
	sessionRepo   domainrepository.SessionRepository
	workspaceRepo domainrepository.WorkspaceRepository
	passwordSvc   auth.PasswordService
	userCloser    UserCloser
}

func New(
	userRepo domainrepository.UserRepository,
	sessionRepo domainrepository.SessionRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	passwordSvc auth.PasswordService,
	userCloser UserCloser,
) *Interactor {
	return &Interactor{userRepo: userRepo, sessionRepo: sessionRepo, workspaceRepo: workspaceRepo, passwordSvc: passwordSvc, userCloser: userCloser}
}

// UpdatePassword は現在のパスワードを確認した上でパスワードを変更し、全セッションを失効させます
func (i *Interactor) UpdatePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	u, err := i.findMe(ctx, userID)
	if err != nil {
		return err
	}
	if err := i.passwordSvc.VerifyPassword(currentPassword, u.PasswordHash); err != nil {
		return domerr.ErrInvalidCredentials
	}
	hashed, err := i.passwordSvc.HashPassword(newPassword)
	if err != nil {
		return err
	}
	u.PasswordHash = hashed

	if err := i.userRepo.Update(ctx, u); err != nil {
		return err
	}

	// パスワード変更後は他端末のセッションも無効化する
	return i.revokeAllSessions(ctx, userID)
}

func (i *Interactor) revokeAllSessions(ctx context.Context, userID string) error {
	if err := i.sessionRepo.RevokeAllByUserID(ctx, userID); err != nil {
		return err
	}
	i.userCloser.CloseUser(userID)
	return nil
}

// DeleteMe は投稿などを残したまま本人のデータを消して退会させます
func (i *Interactor) DeleteMe(ctx context.Context, userID string) error {
	if _, err := i.findMe(ctx, userID); err != nil {
		return err
	}
	memberships, err := i.workspaceRepo.FindMembershipsByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for _, m := range memberships {
		if m.Role == entity.WorkspaceRoleOwner {
			return ErrOwnerCannotDelete
		}
	}

	if err := i.userRepo.Delete(ctx, userID); err != nil {
		return err
	}
	i.userCloser.CloseUser(userID)
	return nil
}

func (i *Interactor) GetMe(ctx context.Context, userID string) (*entity.User, error) {
	return i.findMe(ctx, userID)
}

func (i *Interactor) UpdateMe(ctx context.Context, input UpdateMeInput) (*entity.User, error) {
	u, err := i.findMe(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	if input.DisplayName != nil {
		u.DisplayName = *input.DisplayName
	}
	if input.Bio != nil {
		u.Bio = input.Bio
	}
	// 空文字はアバターを外す
	if input.AvatarURL != nil {
		u.AvatarURL = input.AvatarURL
		if *input.AvatarURL == "" {
			u.AvatarURL = nil
		}
	}

	if input.Links != nil {
		u.Links = *input.Links
	}

	if err := i.userRepo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// UpdatePreferences はテーマ・表示モード・言語・通知の設定を丸ごと置き換えます
func (i *Interactor) UpdatePreferences(ctx context.Context, userID string, preferences entity.UserPreferences) (*entity.UserPreferences, error) {
	u, err := i.findMe(ctx, userID)
	if err != nil {
		return nil, err
	}

	if tz := preferences.Timezone; tz != "" {
		// "Local" はサーバーのタイムゾーンを指すため受け付けない
		if _, err := time.LoadLocation(tz); err != nil || tz == "Local" {
			return nil, fmt.Errorf("%w: %s", domerr.ErrInvalidTimeZone, tz)
		}
	}

	u.Preferences = preferences
	if err := i.userRepo.Update(ctx, u); err != nil {
		return nil, err
	}

	return &u.Preferences, nil
}

func (i *Interactor) findMe(ctx context.Context, userID string) (*entity.User, error) {
	if userID == "" {
		return nil, domerr.ErrUnauthorized
	}

	u, err := i.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil || u.DeletedAt != nil {
		return nil, domerr.ErrUserNotFound
	}
	return u, nil
}
