package reaction

import (
	"context"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/message"
)

type ReactionUseCase interface {
	AddReaction(ctx context.Context, input AddReactionInput) error
	RemoveReaction(ctx context.Context, input RemoveReactionInput) error
	ListReactions(ctx context.Context, messageID string, userID string) (*ListReactionsOutput, error)
}

type reactionInteractor struct {
	messageRepo      domainrepository.MessageRepository
	userRepo         domainrepository.UserRepository
	notificationSvc  Notifier
	channelAccessSvc service.ChannelAccessService
}

func NewReactionInteractor(
	messageRepo domainrepository.MessageRepository,
	userRepo domainrepository.UserRepository,
	notificationSvc Notifier,
	channelAccessSvc service.ChannelAccessService,
) ReactionUseCase {
	return &reactionInteractor{
		messageRepo:      messageRepo,
		userRepo:         userRepo,
		notificationSvc:  notificationSvc,
		channelAccessSvc: channelAccessSvc,
	}
}

// ensureAccess はメッセージがあり、そのチャンネルを閲覧できることを確かめます
func (i *reactionInteractor) ensureAccess(ctx context.Context, messageID, userID string) (*entity.Channel, error) {
	msg, err := i.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch message: %w", err)
	}
	if msg == nil {
		return nil, domerr.ErrMessageNotFound
	}
	return i.channelAccessSvc.EnsureChannelAccess(ctx, msg.ChannelID, userID)
}

// AddReaction は同じリアクションが既にあれば ErrReactionExists を返します
func (i *reactionInteractor) AddReaction(ctx context.Context, input AddReactionInput) error {
	ch, err := i.ensureAccess(ctx, input.MessageID, input.UserID)
	if err != nil {
		return err
	}

	reaction := &entity.MessageReaction{MessageID: input.MessageID, UserID: input.UserID, Emoji: input.Emoji}
	if err := i.messageRepo.AddReaction(ctx, reaction); err != nil {
		return err
	}

	notification := ReactionNotification{MessageID: input.MessageID, UserID: input.UserID, Emoji: input.Emoji, CreatedAt: reaction.CreatedAt}
	if user, err := i.userRepo.FindByID(ctx, input.UserID); err == nil && user != nil {
		notification.User = new(message.NewUserInfo(user))
	}
	i.notificationSvc.NotifyReactionAdded(ch.WorkspaceID, ch.ID, notification)
	return nil
}

func (i *reactionInteractor) RemoveReaction(ctx context.Context, input RemoveReactionInput) error {
	ch, err := i.ensureAccess(ctx, input.MessageID, input.UserID)
	if err != nil {
		return err
	}
	if err := i.messageRepo.RemoveReaction(ctx, input.MessageID, input.UserID, input.Emoji); err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}
	i.notificationSvc.NotifyReactionRemoved(ch.WorkspaceID, ch.ID, ReactionNotification{MessageID: input.MessageID, UserID: input.UserID, Emoji: input.Emoji})
	return nil
}

func (i *reactionInteractor) ListReactions(ctx context.Context, messageID string, userID string) (*ListReactionsOutput, error) {
	if _, err := i.ensureAccess(ctx, messageID, userID); err != nil {
		return nil, err
	}

	reactions, err := i.messageRepo.FindReactions(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reactions: %w", err)
	}

	var userIDs []string
	for _, reaction := range reactions {
		userIDs = append(userIDs, reaction.UserID)
	}
	users, err := i.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}
	userMap := make(map[string]*entity.User, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}

	outputs := make([]ReactionOutput, 0, len(reactions))
	for _, reaction := range reactions {
		outputs = append(outputs, ReactionOutput{
			MessageID: reaction.MessageID,
			User:      message.UserInfoOf(reaction.UserID, userMap),
			Emoji:     reaction.Emoji,
			CreatedAt: reaction.CreatedAt,
		})
	}
	return &ListReactionsOutput{Reactions: outputs}, nil
}
