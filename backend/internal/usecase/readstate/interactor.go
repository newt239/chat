package readstate

import (
	"context"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

type ReadStateUseCase interface {
	GetUnreadCount(ctx context.Context, input GetUnreadCountInput) (*UnreadCountOutput, error)
	UpdateReadState(ctx context.Context, input UpdateReadStateInput) error
}

type readStateInteractor struct {
	readStateRepo    domainrepository.ReadStateRepository
	notificationSvc  Notifier
	channelAccessSvc service.ChannelAccessService
}

func NewReadStateInteractor(
	readStateRepo domainrepository.ReadStateRepository,
	notificationSvc Notifier,
	channelAccessSvc service.ChannelAccessService,
) ReadStateUseCase {
	return &readStateInteractor{
		readStateRepo:    readStateRepo,
		notificationSvc:  notificationSvc,
		channelAccessSvc: channelAccessSvc,
	}
}

func (i *readStateInteractor) GetUnreadCount(ctx context.Context, input GetUnreadCountInput) (*UnreadCountOutput, error) {
	channel, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}

	count, err := i.readStateRepo.GetUnreadCount(ctx, channel.ID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread count: %w", err)
	}

	return &UnreadCountOutput{Count: count}, nil
}

// UpdateReadState は既読位置を保存し、子孫も含めるときは集約表示で見えていた範囲まで子孫を既読にします
func (i *readStateInteractor) UpdateReadState(ctx context.Context, input UpdateReadStateInput) error {
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
	return i.notifyUnreadCounts(ctx, channels, input.UserID)
}

// notifyUnreadCounts は既読にしたチャンネルの未読数とメンション数をまとめて数え、本人の他の端末へ配信します
func (i *readStateInteractor) notifyUnreadCounts(ctx context.Context, channels []*entity.Channel, userID string) error {
	ids := make([]string, len(channels))
	for idx, ch := range channels {
		ids[idx] = ch.ID
	}
	counts, err := i.readStateRepo.GetUnreadCountBatch(ctx, ids, userID)
	if err != nil {
		return fmt.Errorf("failed to get unread counts: %w", err)
	}
	mentions, err := i.readStateRepo.GetUnreadMentionCountBatch(ctx, ids, userID)
	if err != nil {
		return fmt.Errorf("failed to get unread mention counts: %w", err)
	}
	for _, ch := range channels {
		i.notificationSvc.NotifyUnreadCount(ch.WorkspaceID, userID, ch.ID, counts[ch.ID], mentions[ch.ID])
	}
	return nil
}
