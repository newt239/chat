package readstate

import (
	"context"
	"errors"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	domainservice "github.com/newt239/chat/internal/domain/service"
)

var (
	ErrChannelNotFound = errors.New("チャンネルが見つかりません")
	ErrUnauthorized    = errors.New("この操作を行う権限がありません")
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
	notificationSvc   service.NotificationService
	channelAccessSvc  domainservice.ChannelAccessService
}

func NewReadStateInteractor(
	readStateRepo domainrepository.ReadStateRepository,
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	notificationSvc service.NotificationService,
	channelAccessSvc domainservice.ChannelAccessService,
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

	readState := &entity.ChannelReadState{
		ChannelID:  channel.ID,
		UserID:     input.UserID,
		LastReadAt: input.LastReadAt,
	}

	if err := i.readStateRepo.Upsert(ctx, readState); err != nil {
		return fmt.Errorf("failed to update read state: %w", err)
	}

	// 未読数を取得してWebSocket通知を送信（nilチェックを追加）
	if i.notificationSvc != nil {
		count, err := i.readStateRepo.GetUnreadCount(ctx, channel.ID, input.UserID)
		if err == nil {
			i.notificationSvc.NotifyUnreadCount(channel.WorkspaceID, input.UserID, channel.ID, count)
		} else {
			fmt.Printf("Warning: failed to get unread count for notification: %v\n", err)
		}
	}

	return nil
}

// ensureChannelAccess は ChannelAccessService に委譲済み
