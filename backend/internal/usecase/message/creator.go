package message

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/domain/transaction"
)

type MessageCreator struct {
	messageRepo           domainrepository.MessageRepository
	userMentionRepo       domainrepository.MessageUserMentionRepository
	groupMentionRepo      domainrepository.MessageGroupMentionRepository
	linkRepo              domainrepository.MessageLinkRepository
	threadRepo            domainrepository.ThreadRepository
	attachmentRepo        domainrepository.AttachmentRepository
	notificationSvc       Notifier
	mentionService        service.MentionService
	linkProcessingService service.LinkProcessingService
	transactionManager    transaction.Manager
	outputBuilder         *MessageOutputBuilder
	channelAccessSvc      service.ChannelAccessService
	searchIndexer         SearchIndexer
	pushNotifier          PushNotifier
}

func NewMessageCreator(
	messageRepo domainrepository.MessageRepository,
	userMentionRepo domainrepository.MessageUserMentionRepository,
	groupMentionRepo domainrepository.MessageGroupMentionRepository,
	linkRepo domainrepository.MessageLinkRepository,
	threadRepo domainrepository.ThreadRepository,
	attachmentRepo domainrepository.AttachmentRepository,
	notificationSvc Notifier,
	mentionService service.MentionService,
	linkProcessingService service.LinkProcessingService,
	transactionManager transaction.Manager,
	outputBuilder *MessageOutputBuilder,
	channelAccessSvc service.ChannelAccessService,
	searchIndexer SearchIndexer,
	pushNotifier PushNotifier,
) *MessageCreator {
	return &MessageCreator{
		messageRepo:           messageRepo,
		userMentionRepo:       userMentionRepo,
		groupMentionRepo:      groupMentionRepo,
		linkRepo:              linkRepo,
		threadRepo:            threadRepo,
		attachmentRepo:        attachmentRepo,
		notificationSvc:       notificationSvc,
		mentionService:        mentionService,
		linkProcessingService: linkProcessingService,
		transactionManager:    transactionManager,
		outputBuilder:         outputBuilder,
		channelAccessSvc:      channelAccessSvc,
		searchIndexer:         searchIndexer,
		pushNotifier:          pushNotifier,
	}
}

func (c *MessageCreator) CreateMessage(ctx context.Context, input CreateMessageInput) (*MessageOutput, error) {
	if strings.TrimSpace(input.Body) == "" && len(input.AttachmentIDs) == 0 && input.Location == nil {
		return nil, ErrEmptyMessage
	}
	channel, err := c.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}
	if channel.ArchivedAt != nil {
		return nil, domerr.ErrChannelArchived
	}

	if input.ParentID != nil {
		parent, err := c.messageRepo.FindByID(ctx, *input.ParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch parent message: %w", err)
		}
		if parent == nil || parent.ChannelID != channel.ID {
			return nil, ErrParentMessageNotFound
		}
	}

	message := &entity.Message{
		ChannelID: channel.ID,
		UserID:    input.UserID,
		ParentID:  input.ParentID,
		Body:      input.Body,
		Location:  input.Location,
		CreatedAt: time.Now(),
	}
	return c.publish(ctx, channel, message, func(txCtx context.Context) error {
		// スレッド返信は親メッセージの投稿者と返信者を自動フォローする
		if input.ParentID != nil {
			if err := c.followThread(txCtx, *input.ParentID, input.UserID); err != nil {
				return err
			}
		}
		if len(input.AttachmentIDs) > 0 {
			if err := c.verifyAttachments(txCtx, input, channel.ID); err != nil {
				return err
			}
			if err := c.attachmentRepo.AttachToMessage(txCtx, input.AttachmentIDs, message.ID); err != nil {
				return fmt.Errorf("failed to attach files: %w", err)
			}
		}
		return nil
	})
}

// CreateBotMessage はボットユーザー名義のメッセージを投稿します。投稿の可否は呼び出し側で確認済みであることが前提です
func (c *MessageCreator) CreateBotMessage(ctx context.Context, channel *entity.Channel, message *entity.Message) (*MessageOutput, error) {
	message.ChannelID = channel.ID
	message.CreatedAt = time.Now()
	return c.publish(ctx, channel, message, func(context.Context) error { return nil })
}

// publish はメッセージを保存し、メンション・リンクを抽出してチャンネルの購読者へ配信します
func (c *MessageCreator) publish(ctx context.Context, channel *entity.Channel, message *entity.Message, afterCreate func(txCtx context.Context) error) (*MessageOutput, error) {
	var result *MessageOutput
	err := c.transactionManager.Do(ctx, func(txCtx context.Context) error {
		if err := c.messageRepo.Create(txCtx, message); err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}
		if err := afterCreate(txCtx); err != nil {
			return err
		}
		if err := c.extractAndSaveMentionsAndLinks(txCtx, message.ID, message.Body, channel.WorkspaceID); err != nil {
			return fmt.Errorf("failed to extract mentions and links: %w", err)
		}
		outputs, err := c.outputBuilder.Build(txCtx, message.UserID, []*entity.Message{message})
		if err != nil {
			return err
		}
		result = &outputs[0]
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 返信が付くと親メッセージの「スレッドあり」も変わる
	indexIDs := []string{message.ID}
	if message.ParentID != nil {
		indexIDs = append(indexIDs, *message.ParentID)
	}
	c.searchIndexer.Sync(ctx, indexIDs...)

	if c.notificationSvc != nil {
		c.notificationSvc.NotifyNewMessage(channel.WorkspaceID, channel.ID, result.WithoutMessagePreviews())
	}
	c.pushNotifier.NotifyNewMessage(ctx, channel, *result)

	return result, nil
}

func (c *MessageCreator) extractAndSaveMentionsAndLinks(ctx context.Context, messageID, body, workspaceID string) error {
	userMentions, err := c.mentionService.ExtractUserMentions(ctx, body, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to extract user mentions: %w", err)
	}
	for _, mention := range userMentions {
		mention.MessageID = messageID
		mention.CreatedAt = time.Now()
		if err := c.userMentionRepo.Create(ctx, mention); err != nil {
			return fmt.Errorf("failed to create user mention: %w", err)
		}
	}

	groupMentions, err := c.mentionService.ExtractGroupMentions(ctx, body, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to extract group mentions: %w", err)
	}
	for _, mention := range groupMentions {
		mention.MessageID = messageID
		mention.CreatedAt = time.Now()
		if err := c.groupMentionRepo.Create(ctx, mention); err != nil {
			return fmt.Errorf("failed to create group mention: %w", err)
		}
	}

	links, err := c.linkProcessingService.ProcessLinks(ctx, body, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to process links: %w", err)
	}
	for _, link := range links {
		link.MessageID = messageID
		if err := c.linkRepo.Create(ctx, link); err != nil {
			return fmt.Errorf("failed to create link: %w", err)
		}
	}

	return nil
}

// verifyAttachments は添付が投稿者本人のもので、かつ同じチャンネル宛かを検証します
func (c *MessageCreator) verifyAttachments(ctx context.Context, input CreateMessageInput, channelID string) error {
	attachments, err := c.attachmentRepo.FindPendingByIDsForUser(ctx, input.UserID, input.AttachmentIDs)
	if err != nil {
		return fmt.Errorf("failed to verify attachments: %w", err)
	}
	if len(attachments) != len(input.AttachmentIDs) {
		return ErrAttachmentNotFound
	}
	for _, attachment := range attachments {
		if attachment.ChannelID != channelID {
			return ErrAttachmentNotFound
		}
	}
	return nil
}

// followThread は返信者と親メッセージの投稿者をスレッドのフォロワーに登録します
func (c *MessageCreator) followThread(ctx context.Context, threadID, replierID string) error {
	parent, err := c.messageRepo.FindByID(ctx, threadID)
	if err != nil {
		return fmt.Errorf("failed to fetch parent message: %w", err)
	}
	if parent == nil {
		return ErrParentMessageNotFound
	}

	for _, userID := range []string{parent.UserID, replierID} {
		if err := c.threadRepo.FollowThread(ctx, userID, threadID); err != nil {
			return fmt.Errorf("failed to follow thread: %w", err)
		}
	}
	return nil
}
