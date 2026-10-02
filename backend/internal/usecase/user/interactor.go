package user

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainerrors "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/usecase/auth"
)

var (
	ErrUnauthorized    = errors.New("この操作を行う権限がありません")
	ErrInvalidTimeZone = errors.New("タイムゾーンの指定が正しくありません")
	ErrInvalidLink     = fmt.Errorf("%w: リンクは %d 件までの http(s) の URL で指定してください", domainerrors.ErrValidation, entity.MaxProfileLinks)
)

type UseCase interface {
	GetMe(ctx context.Context, userID string) (*MeOutput, error)
	UpdateMe(ctx context.Context, input UpdateMeInput) (*MeOutput, error)
	UpdatePreferences(ctx context.Context, input UpdatePreferencesInput) (*entity.UserPreferences, error)
	UpdatePassword(ctx context.Context, input UpdatePasswordInput) error
	DeleteMe(ctx context.Context, userID string) error
}

// UserCloser は全セッションを失効させたユーザーのリアルタイム接続を切ります
type UserCloser interface {
	CloseUser(userID string)
}

type interactor struct {
	userRepo    domainrepository.UserRepository
	sessionRepo domainrepository.SessionRepository
	passwordSvc auth.PasswordService
	userCloser  UserCloser
}

func NewInteractor(
	userRepo domainrepository.UserRepository,
	sessionRepo domainrepository.SessionRepository,
	passwordSvc auth.PasswordService,
	userCloser UserCloser,
) UseCase {
	return &interactor{userRepo: userRepo, sessionRepo: sessionRepo, passwordSvc: passwordSvc, userCloser: userCloser}
}

// UpdatePassword は現在のパスワードを確認した上でパスワードを変更し、全セッションを失効させます
func (i *interactor) UpdatePassword(ctx context.Context, input UpdatePasswordInput) error {
	u, err := i.findMe(ctx, input.UserID)
	if err != nil {
		return err
	}

	if err := i.passwordSvc.VerifyPassword(input.CurrentPassword, u.PasswordHash); err != nil {
		return domainerrors.ErrInvalidCredentials
	}

	hashed, err := i.passwordSvc.HashPassword(input.NewPassword)
	if err != nil {
		return err
	}
	u.PasswordHash = hashed

	if err := i.userRepo.Update(ctx, u); err != nil {
		return err
	}

	// パスワード変更後は他端末のセッションも無効化する
	return i.revokeAllSessions(ctx, input.UserID)
}

func (i *interactor) revokeAllSessions(ctx context.Context, userID string) error {
	if err := i.sessionRepo.RevokeAllByUserID(ctx, userID); err != nil {
		return err
	}
	i.userCloser.CloseUser(userID)
	return nil
}

// DeleteMe はアカウントを削除します
func (i *interactor) DeleteMe(ctx context.Context, userID string) error {
	if _, err := i.findMe(ctx, userID); err != nil {
		return err
	}

	if err := i.revokeAllSessions(ctx, userID); err != nil {
		return err
	}

	return i.userRepo.Delete(ctx, userID)
}

func (i *interactor) GetMe(ctx context.Context, userID string) (*MeOutput, error) {
	u, err := i.findMe(ctx, userID)
	if err != nil {
		return nil, err
	}

	return toMeOutput(u), nil
}

func (i *interactor) UpdateMe(ctx context.Context, input UpdateMeInput) (*MeOutput, error) {
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
		links, err := normalizeLinks(*input.Links)
		if err != nil {
			return nil, err
		}
		u.Links = links
	}

	if err := i.userRepo.Update(ctx, u); err != nil {
		return nil, err
	}

	return toMeOutput(u), nil
}

// normalizeLinks は前後の空白を除き、数と URL の形式を確かめます
func normalizeLinks(links []string) ([]string, error) {
	if len(links) > entity.MaxProfileLinks {
		return nil, ErrInvalidLink
	}
	result := make([]string, 0, len(links))
	for _, link := range links {
		link = strings.TrimSpace(link)
		u, err := url.Parse(link)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return nil, ErrInvalidLink
		}
		result = append(result, link)
	}
	return result, nil
}

// UpdatePreferences はテーマ・表示モード・言語・通知の設定を丸ごと置き換えます
func (i *interactor) UpdatePreferences(ctx context.Context, input UpdatePreferencesInput) (*entity.UserPreferences, error) {
	u, err := i.findMe(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	if tz := input.Preferences.Timezone; tz != "" {
		// "Local" はサーバーのタイムゾーンを指すため受け付けない
		if _, err := time.LoadLocation(tz); err != nil || tz == "Local" {
			return nil, fmt.Errorf("%w: %s", ErrInvalidTimeZone, tz)
		}
	}

	u.Preferences = input.Preferences
	if err := i.userRepo.Update(ctx, u); err != nil {
		return nil, err
	}

	return &u.Preferences, nil
}

func (i *interactor) findMe(ctx context.Context, userID string) (*entity.User, error) {
	if userID == "" {
		return nil, ErrUnauthorized
	}

	u, err := i.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, entity.ErrUserNotFound
	}
	return u, nil
}

func toMeOutput(u *entity.User) *MeOutput {
	return &MeOutput{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Bio:         u.Bio,
		AvatarURL:   u.AvatarURL,
		Links:       u.Links,
		Preferences: u.Preferences,
	}
}
