package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type Interactor struct {
	pushTokenRepo domainrepository.PushTokenRepository
}

func New(pushTokenRepo domainrepository.PushTokenRepository) *Interactor {
	return &Interactor{pushTokenRepo: pushTokenRepo}
}

// RegisterPushToken は端末のトークンを登録します。起動のたびに呼ばれ、最終利用日時を更新します
func (i *Interactor) RegisterPushToken(ctx context.Context, token entity.PushToken) error {
	token.LastSeenAt = time.Now()
	if err := i.pushTokenRepo.Upsert(ctx, &token); err != nil {
		return fmt.Errorf("failed to register push token: %w", err)
	}
	return nil
}

// UnregisterPushToken は自分のトークンだけを削除します。登録がなくても成功とします
func (i *Interactor) UnregisterPushToken(ctx context.Context, userID, token string) error {
	if err := i.pushTokenRepo.Delete(ctx, userID, token); err != nil {
		return fmt.Errorf("failed to unregister push token: %w", err)
	}
	return nil
}
