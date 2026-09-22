package user

import (
	"context"
	"errors"

	"github.com/newt239/chat/internal/domain/entity"
	domainerrors "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/usecase/auth"
)

var (
	ErrUnauthorized = errors.New("この操作を行う権限がありません")
)

type UseCase interface {
	GetMe(ctx context.Context, userID string) (*MeOutput, error)
	UpdateMe(ctx context.Context, input UpdateMeInput) (*MeOutput, error)
	UpdatePassword(ctx context.Context, input UpdatePasswordInput) error
	DeleteMe(ctx context.Context, userID string) error
}

type interactor struct {
	userRepo    domainrepository.UserRepository
	sessionRepo domainrepository.SessionRepository
	passwordSvc auth.PasswordService
}

func NewInteractor(
	userRepo domainrepository.UserRepository,
	sessionRepo domainrepository.SessionRepository,
	passwordSvc auth.PasswordService,
) UseCase {
	return &interactor{userRepo: userRepo, sessionRepo: sessionRepo, passwordSvc: passwordSvc}
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
	return i.sessionRepo.RevokeAllByUserID(ctx, input.UserID)
}

// DeleteMe はアカウントを削除します
func (i *interactor) DeleteMe(ctx context.Context, userID string) error {
	if _, err := i.findMe(ctx, userID); err != nil {
		return err
	}

	if err := i.sessionRepo.RevokeAllByUserID(ctx, userID); err != nil {
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
	if input.AvatarURL != nil {
		u.AvatarURL = input.AvatarURL
	}

	if err := i.userRepo.Update(ctx, u); err != nil {
		return nil, err
	}

	return toMeOutput(u), nil
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
	}
}
