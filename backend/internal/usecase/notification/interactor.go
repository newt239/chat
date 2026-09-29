package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type RegisterPushTokenInput struct {
	UserID    string
	Token     string
	Platform  entity.PushPlatform
	UserAgent string
}

type UnregisterPushTokenInput struct {
	UserID string
	Token  string
}

type UseCase interface {
	RegisterPushToken(ctx context.Context, input RegisterPushTokenInput) error
	UnregisterPushToken(ctx context.Context, input UnregisterPushTokenInput) error
}

type interactor struct {
	pushTokenRepo domainrepository.PushTokenRepository
}

func NewInteractor(pushTokenRepo domainrepository.PushTokenRepository) UseCase {
	return &interactor{pushTokenRepo: pushTokenRepo}
}

// RegisterPushToken は端末のトークンを登録します。起動のたびに呼ばれ、最終利用日時を更新します
func (i *interactor) RegisterPushToken(ctx context.Context, input RegisterPushTokenInput) error {
	err := i.pushTokenRepo.Upsert(ctx, &entity.PushToken{
		UserID:     input.UserID,
		Token:      input.Token,
		Platform:   input.Platform,
		UserAgent:  input.UserAgent,
		LastSeenAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to register push token: %w", err)
	}
	return nil
}

// UnregisterPushToken は自分のトークンだけを削除します。登録がなくても成功とします
func (i *interactor) UnregisterPushToken(ctx context.Context, input UnregisterPushTokenInput) error {
	if err := i.pushTokenRepo.Delete(ctx, input.UserID, input.Token); err != nil {
		return fmt.Errorf("failed to unregister push token: %w", err)
	}
	return nil
}
