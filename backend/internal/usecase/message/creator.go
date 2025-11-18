package message

import (
	"context"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/domain/transaction"
)

type MessageCreator struct {
	messageRepo           domainrepository.MessageRepository
	channelRepo           domainrepository.ChannelRepository
	channelMemberRepo     domainrepository.ChannelMemberRepository
	workspaceRepo         domainrepository.WorkspaceRepository
	userRepo              domainrepository.UserRepository
	userGroupRepo         domainrepository.UserGroupRepository
	userMentionRepo       domainrepository.MessageUserMentionRepository
	groupMentionRepo      domainrepository.MessageGroupMentionRepository
	linkRepo              domainrepository.MessageLinkRepository
	threadRepo            domainrepository.ThreadRepository
	attachmentRepo        domainrepository.AttachmentRepository
	ogpService            service.OGPService
	notificationSvc       service.NotificationService
	mentionService        service.MentionService
	linkProcessingService service.LinkProcessingService
	transactionManager    transaction.Manager
	assembler             *MessageOutputAssembler
    channelAccessSvc      service.ChannelAccessService
}

func NewMessageCreator(
	messageRepo domainrepository.MessageRepository,
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
	userGroupRepo domainrepository.UserGroupRepository,
	userMentionRepo domainrepository.MessageUserMentionRepository,
	groupMentionRepo domainrepository.MessageGroupMentionRepository,
	linkRepo domainrepository.MessageLinkRepository,
	threadRepo domainrepository.ThreadRepository,
	attachmentRepo domainrepository.AttachmentRepository,
	ogpService service.OGPService,
	notificationSvc service.NotificationService,
	mentionService service.MentionService,
	linkProcessingService service.LinkProcessingService,
	transactionManager transaction.Manager,
    channelAccessSvc service.ChannelAccessService,
) *MessageCreator {
	return &MessageCreator{
		messageRepo:           messageRepo,
		channelRepo:           channelRepo,
		channelMemberRepo:     channelMemberRepo,
		workspaceRepo:         workspaceRepo,
		userRepo:              userRepo,
		userGroupRepo:         userGroupRepo,
		userMentionRepo:       userMentionRepo,
		groupMentionRepo:      groupMentionRepo,
		linkRepo:              linkRepo,
		threadRepo:            threadRepo,
		attachmentRepo:        attachmentRepo,
		ogpService:            ogpService,
		notificationSvc:       notificationSvc,
		mentionService:        mentionService,
		linkProcessingService: linkProcessingService,
		transactionManager:    transactionManager,
		assembler:             NewMessageOutputAssembler(),
        channelAccessSvc:      channelAccessSvc,
	}
}

func (c *MessageCreator) CreateMessage(ctx context.Context, input CreateMessageInput) (*MessageOutput, error) {
    channel, err := c.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
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

	var result *MessageOutput
	err = c.transactionManager.Do(ctx, func(txCtx context.Context) error {
		message := &entity.Message{
			ChannelID: channel.ID,
			UserID:    input.UserID,
			ParentID:  input.ParentID,
			Body:      input.Body,
			CreatedAt: time.Now(),
		}

		if err := c.messageRepo.Create(txCtx, message); err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}

		if len(input.AttachmentIDs) > 0 {
			if err := c.attachmentRepo.AttachToMessage(txCtx, input.AttachmentIDs, message.ID); err != nil {
				return fmt.Errorf("failed to attach files: %w", err)
			}
		}

		if err := c.extractAndSaveMentionsAndLinks(txCtx, message.ID, input.Body, channel.WorkspaceID); err != nil {
			return fmt.Errorf("failed to extract mentions and links: %w", err)
		}

		user, err := c.userRepo.FindByID(txCtx, input.UserID)
		if err != nil {
			return fmt.Errorf("failed to fetch user: %w", err)
		}

		userMentions, err := c.userMentionRepo.FindByMessageID(txCtx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch user mentions: %w", err)
		}
		groupMentions, err := c.groupMentionRepo.FindByMessageID(txCtx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch group mentions: %w", err)
		}
		links, err := c.linkRepo.FindByMessageID(txCtx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch links: %w", err)
		}
		attachmentList, err := c.attachmentRepo.FindByMessageID(txCtx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch attachments: %w", err)
		}

		groupIDs := make([]string, 0)
		groupIDSet := make(map[string]bool)
		for _, mention := range groupMentions {
			if !groupIDSet[mention.GroupID] {
				groupIDs = append(groupIDs, mention.GroupID)
				groupIDSet[mention.GroupID] = true
			}
		}

		groups := make(map[string]*entity.UserGroup)
		if len(groupIDs) > 0 {
			groupList, err := c.userGroupRepo.FindByIDs(txCtx, groupIDs)
			if err != nil {
				return fmt.Errorf("failed to fetch groups: %w", err)
			}
			for _, group := range groupList {
				groups[group.ID] = group
			}
		}

		reactions := []*entity.MessageReaction{}

		userMap := map[string]*entity.User{user.ID: user}

		output := c.assembler.AssembleMessageOutput(message, user, userMentions, groupMentions, links, reactions, attachmentList, groups, userMap)
		result = &output

		return nil
	})

	if err != nil {
		return nil, err
	}

	if c.notificationSvc != nil {
		c.notificationSvc.NotifyNewMessage(channel.WorkspaceID, channel.ID, *result)
	}

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

	links, err := c.linkProcessingService.ProcessLinks(ctx, body)
	if err != nil {
		return fmt.Errorf("failed to process links: %w", err)
	}

	for _, link := range links {
		existingLink, err := c.linkRepo.FindByURL(ctx, link.URL)
		if err != nil {
			continue // エラーは無視
		}

		if existingLink != nil {
			link.MessageID = messageID
			link.Title = existingLink.Title
			link.Description = existingLink.Description
			link.ImageURL = existingLink.ImageURL
			link.SiteName = existingLink.SiteName
			link.CardType = existingLink.CardType
			link.CreatedAt = time.Now()
		} else {
			link.MessageID = messageID
			link.CreatedAt = time.Now()
		}

		if err := c.linkRepo.Create(ctx, link); err != nil {
			return fmt.Errorf("failed to create link: %w", err)
		}
	}

	return nil
}
