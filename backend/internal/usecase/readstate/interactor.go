package readstate

import (
	"context"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

type ReadStateUseCase interface {
	GetUnreadCount(ctx context.Context, input GetUnreadCountInput) (*UnreadCountOutput, error)
	UpdateReadState(ctx context.Context, input UpdateReadStateInput) error
}

type readStateInteractor struct {
	readStateRepo     domainrepository.ReadStateRepository
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	workspaceRepo     domainrepository.WorkspaceRepository
	notificationSvc   Notifier
	channelAccessSvc  service.ChannelAccessService
}

func NewReadStateInteractor(
	readStateRepo domainrepository.ReadStateRepository,
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	notificationSvc Notifier,
	channelAccessSvc service.ChannelAccessService,
) ReadStateUseCase {
	return &readStateInteractor{
		readStateRepo:     readStateRepo,
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		workspaceRepo:     workspaceRepo,
		notificationSvc:   notificationSvc,
		channelAccessSvc:  channelAccessSvc,
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

func (i *readStateInteractor) UpdateReadState(ctx context.Context, input UpdateReadStateInput) error {
	channel, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return err
	}

	if err := i.markAsRead(ctx, channel, input.UserID, input.LastReadAt); err != nil {
		return err
	}
	if !input.IncludeDescendants {
		return nil
	}

	descendants, err := i.channelAccessSvc.AccessibleDescendants(ctx, channel, input.UserID)
	if err != nil {
		return err
	}
	for _, ch := range descendants {
		// 集約表示で見えていた範囲までを既読にするため、既読位置は戻さない
		current, err := i.readStateRepo.FindByChannelAndUser(ctx, ch.ID, input.UserID)
		if err != nil {
			return fmt.Errorf("failed to load read state: %w", err)
		}
		if current != nil && !current.LastReadAt.Before(input.LastReadAt) {
			continue
		}
		if err := i.markAsRead(ctx, ch, input.UserID, input.LastReadAt); err != nil {
			return err
		}
	}
	return nil
}

func (i *readStateInteractor) markAsRead(ctx context.Context, channel *entity.Channel, userID string, lastReadAt time.Time) error {
	readState := &entity.ChannelReadState{
		ChannelID:  channel.ID,
		UserID:     userID,
		LastReadAt: lastReadAt,
	}

	if err := i.readStateRepo.Upsert(ctx, readState); err != nil {
		return fmt.Errorf("failed to update read state: %w", err)
	}

	// 未読数とメンション数を取得してWebSocket通知を送信
	if i.notificationSvc != nil {
		count, err := i.readStateRepo.GetUnreadCount(ctx, channel.ID, userID)
		if err != nil {
			fmt.Printf("Warning: failed to get unread count for notification: %v\n", err)
			return nil
		}

		mentionCounts, err := i.readStateRepo.GetUnreadMentionCountBatch(ctx, []string{channel.ID}, userID)
		if err != nil {
			fmt.Printf("Warning: failed to get unread mention count for notification: %v\n", err)
		}

		i.notificationSvc.NotifyUnreadCount(channel.WorkspaceID, userID, channel.ID, count, mentionCounts[channel.ID])
	}

	return nil
}
