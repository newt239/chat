package user

import (
	"context"
	"errors"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

var (
	ErrUnauthorized = errors.New("この操作を行う権限がありません")
)

type UseCase interface {
	GetMe(ctx context.Context, userID string) (*MeOutput, error)
	UpdateMe(ctx context.Context, input UpdateMeInput) (*MeOutput, error)
}

type interactor struct {
	userRepo domainrepository.UserRepository
}

func NewInteractor(userRepo domainrepository.UserRepository) UseCase {
	return &interactor{userRepo: userRepo}
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
