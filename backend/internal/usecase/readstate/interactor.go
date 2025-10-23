package readstate

import (
	"errors"
	"fmt"

	"github.com/example/chat/internal/domain"
)

var (
	ErrChannelNotFound = errors.New("channel not found")
	ErrUnauthorized    = errors.New("unauthorized to perform this action")
)

type ReadStateUseCase interface {
	GetUnreadCount(input GetUnreadCountInput) (*UnreadCountOutput, error)
	UpdateReadState(input UpdateReadStateInput) error
}

type readStateInteractor struct {
	readStateRepo domain.ReadStateRepository
	channelRepo   domain.ChannelRepository
	workspaceRepo domain.WorkspaceRepository
}

func NewReadStateInteractor(
	readStateRepo domain.ReadStateRepository,
	channelRepo domain.ChannelRepository,
	workspaceRepo domain.WorkspaceRepository,
) ReadStateUseCase {
	return &readStateInteractor{
		readStateRepo: readStateRepo,
		channelRepo:   channelRepo,
		workspaceRepo: workspaceRepo,
	}
}

func (i *readStateInteractor) GetUnreadCount(input GetUnreadCountInput) (*UnreadCountOutput, error) {
	channel, err := i.ensureChannelAccess(input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}

	count, err := i.readStateRepo.GetUnreadCount(channel.ID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread count: %w", err)
	}

	return &UnreadCountOutput{Count: count}, nil
}

func (i *readStateInteractor) UpdateReadState(input UpdateReadStateInput) error {
	channel, err := i.ensureChannelAccess(input.ChannelID, input.UserID)
	if err != nil {
		return err
	}

	readState := &domain.ChannelReadState{
		ChannelID:  channel.ID,
		UserID:     input.UserID,
		LastReadAt: input.LastReadAt,
	}

	if err := i.readStateRepo.Upsert(readState); err != nil {
		return fmt.Errorf("failed to update read state: %w", err)
	}

	return nil
}

func (i *readStateInteractor) ensureChannelAccess(channelID, userID string) (*domain.Channel, error) {
	ch, err := i.channelRepo.FindByID(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to load channel: %w", err)
	}
	if ch == nil {
		return nil, ErrChannelNotFound
	}

	if ch.IsPrivate {
		isMember, err := i.channelRepo.IsMember(ch.ID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify channel membership: %w", err)
		}
		if !isMember {
			return nil, ErrUnauthorized
		}
		return ch, nil
	}

	member, err := i.workspaceRepo.FindMember(ch.WorkspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	return ch, nil
}
