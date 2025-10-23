package message

import (
	"errors"
	"fmt"
	"time"

	"github.com/example/chat/internal/domain"
)

var (
	ErrChannelNotFound       = errors.New("channel not found")
	ErrUnauthorized          = errors.New("unauthorized to perform this action")
	ErrParentMessageNotFound = errors.New("parent message not found")
)

const (
	defaultMessageLimit = 50
	maxMessageLimit     = 100
)

type MessageUseCase interface {
	ListMessages(input ListMessagesInput) (*ListMessagesOutput, error)
	CreateMessage(input CreateMessageInput) (*MessageOutput, error)
}

type messageInteractor struct {
	messageRepo   domain.MessageRepository
	channelRepo   domain.ChannelRepository
	workspaceRepo domain.WorkspaceRepository
}

func NewMessageInteractor(
	messageRepo domain.MessageRepository,
	channelRepo domain.ChannelRepository,
	workspaceRepo domain.WorkspaceRepository,
) MessageUseCase {
	return &messageInteractor{
		messageRepo:   messageRepo,
		channelRepo:   channelRepo,
		workspaceRepo: workspaceRepo,
	}
}

func (i *messageInteractor) ListMessages(input ListMessagesInput) (*ListMessagesOutput, error) {
	channel, err := i.ensureChannelAccess(input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}

	limit := input.Limit
	if limit <= 0 {
		limit = defaultMessageLimit
	}
	if limit > maxMessageLimit {
		limit = maxMessageLimit
	}

	fetchLimit := limit + 1

	messages, err := i.messageRepo.FindByChannelID(channel.ID, fetchLimit, input.Since, input.Until)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	hasMore := false
	if len(messages) > limit {
		hasMore = true
		messages = messages[:limit]
	}

	outputs := make([]MessageOutput, 0, len(messages))
	for _, msg := range messages {
		outputs = append(outputs, toMessageOutput(msg))
	}

	return &ListMessagesOutput{Messages: outputs, HasMore: hasMore}, nil
}

func (i *messageInteractor) CreateMessage(input CreateMessageInput) (*MessageOutput, error) {
	channel, err := i.ensureChannelAccess(input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}

	if input.ParentID != nil {
		parent, err := i.messageRepo.FindByID(*input.ParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch parent message: %w", err)
		}
		if parent == nil || parent.ChannelID != channel.ID {
			return nil, ErrParentMessageNotFound
		}
	}

	message := &domain.Message{
		ChannelID: channel.ID,
		UserID:    input.UserID,
		ParentID:  input.ParentID,
		Body:      input.Body,
		CreatedAt: time.Now(),
	}

	if err := i.messageRepo.Create(message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	output := toMessageOutput(message)
	return &output, nil
}

func (i *messageInteractor) ensureChannelAccess(channelID, userID string) (*domain.Channel, error) {
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

func toMessageOutput(message *domain.Message) MessageOutput {
	return MessageOutput{
		ID:        message.ID,
		ChannelID: message.ChannelID,
		UserID:    message.UserID,
		ParentID:  message.ParentID,
		Body:      message.Body,
		CreatedAt: message.CreatedAt,
		EditedAt:  message.EditedAt,
		DeletedAt: message.DeletedAt,
	}
}
