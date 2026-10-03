package reaction

import (
	"context"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/message"
)

type ReactionInput struct {
	MessageID string
	UserID    string
	Emoji     string
}

type Interactor struct {
	messageRepo      domainrepository.MessageRepository
	userRepo         domainrepository.UserRepository
	notificationSvc  Notifier
	channelAccessSvc service.ChannelAccessService
}

func New(
	messageRepo domainrepository.MessageRepository,
	userRepo domainrepository.UserRepository,
	notificationSvc Notifier,
	channelAccessSvc service.ChannelAccessService,
) *Interactor {
	return &Interactor{messageRepo: messageRepo, userRepo: userRepo, notificationSvc: notificationSvc, channelAccessSvc: channelAccessSvc}
}

// AddReaction は同じリアクションが既にあれば ErrReactionExists を返します
func (i *Interactor) AddReaction(ctx context.Context, input ReactionInput) error {
	_, ch, err := i.channelAccessSvc.EnsureMessageAccess(ctx, input.MessageID, input.UserID)
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

func (i *Interactor) RemoveReaction(ctx context.Context, input ReactionInput) error {
	_, ch, err := i.channelAccessSvc.EnsureMessageAccess(ctx, input.MessageID, input.UserID)
	if err != nil {
		return err
	}
	if err := i.messageRepo.RemoveReaction(ctx, input.MessageID, input.UserID, input.Emoji); err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}
	i.notificationSvc.NotifyReactionRemoved(ch.WorkspaceID, ch.ID, ReactionNotification{MessageID: input.MessageID, UserID: input.UserID, Emoji: input.Emoji})
	return nil
}
