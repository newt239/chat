package message

import (
	"context"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/domain/transaction"
)

type Interactor struct {
	messageRepo      domainrepository.MessageRepository
	systemMsgRepo    domainrepository.SystemMessageRepository
	userRepo         domainrepository.UserRepository
	workspaceRepo    domainrepository.WorkspaceRepository
	threadRepo       domainrepository.ThreadRepository
	attachmentRepo   domainrepository.AttachmentRepository
	pollRepo         domainrepository.PollRepository
	mentionRepo      domainrepository.MessageMentionRepository
	linkRepo         domainrepository.MessageLinkRepository
	mentionSvc       service.MentionService
	linkSvc          *service.LinkProcessingService
	ogpSvc           service.OGPService
	channelAccessSvc service.ChannelAccessService
	permissionSvc    service.PermissionService
	txManager        transaction.Manager
	outputBuilder    *MessageOutputBuilder
	notifier         Notifier
	searchIndexer    SearchIndexer
	observers        []NewMessageObserver
}

func New(
	messageRepo domainrepository.MessageRepository,
	systemMsgRepo domainrepository.SystemMessageRepository,
	userRepo domainrepository.UserRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	threadRepo domainrepository.ThreadRepository,
	attachmentRepo domainrepository.AttachmentRepository,
	pollRepo domainrepository.PollRepository,
	mentionRepo domainrepository.MessageMentionRepository,
	linkRepo domainrepository.MessageLinkRepository,
	mentionSvc service.MentionService,
	linkSvc *service.LinkProcessingService,
	ogpSvc service.OGPService,
	channelAccessSvc service.ChannelAccessService,
	permissionSvc service.PermissionService,
	txManager transaction.Manager,
	outputBuilder *MessageOutputBuilder,
	notifier Notifier,
	searchIndexer SearchIndexer,
	observers []NewMessageObserver,
) *Interactor {
	return &Interactor{
		messageRepo:      messageRepo,
		systemMsgRepo:    systemMsgRepo,
		userRepo:         userRepo,
		workspaceRepo:    workspaceRepo,
		threadRepo:       threadRepo,
		attachmentRepo:   attachmentRepo,
		pollRepo:         pollRepo,
		mentionRepo:      mentionRepo,
		linkRepo:         linkRepo,
		mentionSvc:       mentionSvc,
		linkSvc:          linkSvc,
		ogpSvc:           ogpSvc,
		channelAccessSvc: channelAccessSvc,
		permissionSvc:    permissionSvc,
		txManager:        txManager,
		outputBuilder:    outputBuilder,
		notifier:         notifier,
		searchIndexer:    searchIndexer,
		observers:        observers,
	}
}

// FetchOGP は投稿前のプレビューに使う OGP を取得します
func (i *Interactor) FetchOGP(ctx context.Context, url string) (*entity.OGPData, error) {
	ogp, err := i.ogpSvc.FetchOGP(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OGP data: %w", err)
	}
	return ogp, nil
}

// EnsureReplyTarget は parentID があれば、それが channelID のスレッドの親にできるメッセージかを確かめます
func EnsureReplyTarget(ctx context.Context, messageRepo domainrepository.MessageRepository, parentID *string, channelID string) (*entity.Message, error) {
	if parentID == nil {
		return nil, nil
	}
	parent, err := messageRepo.FindByID(ctx, *parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load parent message: %w", err)
	}
	if !parent.CanBeRepliedIn(channelID) {
		return nil, domerr.ErrParentMessageNotFound
	}
	return parent, nil
}

// VerifyAttachments は添付が投稿者本人のまだ使っていないもので、同じチャンネル宛てかを確かめます
func VerifyAttachments(ctx context.Context, attachmentRepo domainrepository.AttachmentRepository, userID, channelID string, attachmentIDs []string) error {
	if len(attachmentIDs) == 0 {
		return nil
	}
	attachments, err := attachmentRepo.FindPendingByIDsForUser(ctx, userID, attachmentIDs)
	if err != nil {
		return fmt.Errorf("failed to verify attachments: %w", err)
	}
	if len(attachments) != len(attachmentIDs) {
		return domerr.ErrAttachmentNotFound
	}
	for _, attachment := range attachments {
		if attachment.ChannelID != channelID {
			return domerr.ErrAttachmentNotFound
		}
	}
	return nil
}
