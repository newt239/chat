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
    UpdateMe(ctx context.Context, input UpdateMeInput) (*UpdateMeOutput, error)
}

type interactor struct {
    userRepo domainrepository.UserRepository
}

func NewInteractor(userRepo domainrepository.UserRepository) UseCase {
    return &interactor{userRepo: userRepo}
}

func (i *interactor) UpdateMe(ctx context.Context, input UpdateMeInput) (*UpdateMeOutput, error) {
    if input.UserID == "" {
        return nil, ErrUnauthorized
    }

    u, err := i.userRepo.FindByID(ctx, input.UserID)
    if err != nil {
        return nil, err
    }
    if u == nil {
        return nil, entity.ErrUserNotFound
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

    return &UpdateMeOutput{
        ID:          u.ID,
        DisplayName: u.DisplayName,
        Bio:         u.Bio,
        AvatarURL:   u.AvatarURL,
    }, nil
}


