package readstate

import (
	"context"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

type UpdateReadStateInput struct {
	ChannelID          string
	UserID             string
	LastReadAt         time.Time
	IncludeDescendants bool
}

type Interactor struct {
	readStateRepo    domainrepository.ReadStateRepository
	notificationSvc  Notifier
	channelAccessSvc service.ChannelAccessService
}

func New(readStateRepo domainrepository.ReadStateRepository, notificationSvc Notifier, channelAccessSvc service.ChannelAccessService) *Interactor {
	return &Interactor{readStateRepo: readStateRepo, notificationSvc: notificationSvc, channelAccessSvc: channelAccessSvc}
}

// UpdateReadState は既読位置を保存し、子孫も含めるときは集約表示で見えていた範囲まで子孫を既読にします
func (i *Interactor) UpdateReadState(ctx context.Context, input UpdateReadStateInput) error {
	channel, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return err
	}
	if err := i.readStateRepo.Upsert(ctx, &entity.ChannelReadState{ChannelID: channel.ID, UserID: input.UserID, LastReadAt: input.LastReadAt}); err != nil {
		return fmt.Errorf("failed to update read state: %w", err)
	}

	channels := []*entity.Channel{channel}
	if input.IncludeDescendants {
		descendants, err := i.channelAccessSvc.AccessibleDescendants(ctx, channel, input.UserID)
		if err != nil {
			return err
		}
		ids := make([]string, len(descendants))
		for idx, ch := range descendants {
			ids[idx] = ch.ID
		}
		// 子孫は既読位置を戻さない
		if err := i.readStateRepo.AdvanceBatch(ctx, ids, input.UserID, input.LastReadAt); err != nil {
			return fmt.Errorf("failed to update descendant read states: %w", err)
		}
		channels = append(channels, descendants...)
	}

	// 既読にしたチャンネルの未読数とメンション数を本人の他の端末へ配信する
	ids := make([]string, len(channels))
	for idx, ch := range channels {
		ids[idx] = ch.ID
	}
	counts, err := i.readStateRepo.GetUnreadCountBatch(ctx, ids, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to get unread counts: %w", err)
	}
	mentions, err := i.readStateRepo.GetUnreadMentionCountBatch(ctx, ids, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to get unread mention counts: %w", err)
	}
	for _, ch := range channels {
		i.notificationSvc.NotifyUnreadCount(ch.WorkspaceID, input.UserID, ch.ID, counts[ch.ID], mentions[ch.ID])
	}
	return nil
}
