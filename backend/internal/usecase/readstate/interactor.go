package readstate

import (
	"context"
	"errors"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

var (
	ErrChannelNotFound = errors.New("channel not found")
	ErrUnauthorized    = errors.New("unauthorized to perform this action")
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
}

func NewReadStateInteractor(
	readStateRepo domainrepository.ReadStateRepository,
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	notificationSvc service.NotificationService,
) ReadStateUseCase {
	return &readStateInteractor{
		readStateRepo:     readStateRepo,
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		workspaceRepo:     workspaceRepo,
		notificationSvc:   notificationSvc,
	}
}

func (i *readStateInteractor) GetUnreadCount(ctx context.Context, input GetUnreadCountInput) (*UnreadCountOutput, error) {
	channel, err := i.ensureChannelAccess(ctx, input.ChannelID, input.UserID)
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
	channel, err := i.ensureChannelAccess(ctx, input.ChannelID, input.UserID)
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

func (i *readStateInteractor) ensureChannelAccess(ctx context.Context, channelID, userID string) (*entity.Channel, error) {
	ch, err := i.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to load channel: %w", err)
	}
	if ch == nil {
		return nil, ErrChannelNotFound
	}

	if ch.IsPrivate {
		isMember, err := i.channelMemberRepo.IsMember(ctx, ch.ID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify channel membership: %w", err)
		}
		if !isMember {
			return nil, ErrUnauthorized
		}
		return ch, nil
	}

	member, err := i.workspaceRepo.FindMember(ctx, ch.WorkspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	return ch, nil
}
