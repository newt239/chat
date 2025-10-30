package bookmark

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/message"
)

var (
	ErrMessageNotFound = errors.New("メッセージが見つかりません")
	ErrUnauthorized    = errors.New("この操作を行う権限がありません")
	ErrBookmarkExists  = errors.New("このメッセージは既にブックマークされています")
)

type BookmarkUseCase interface {
	AddBookmark(ctx context.Context, input AddBookmarkInput) error
	RemoveBookmark(ctx context.Context, input RemoveBookmarkInput) error
	ListBookmarks(ctx context.Context, userID string) (*ListBookmarksOutput, error)
	IsBookmarked(ctx context.Context, userID, messageID string) (bool, error)
}

type bookmarkInteractor struct {
	bookmarkRepo      domainrepository.BookmarkRepository
	messageRepo       domainrepository.MessageRepository
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	workspaceRepo     domainrepository.WorkspaceRepository
	userRepo          domainrepository.UserRepository
	mentionRepo       domainrepository.MessageUserMentionRepository
	groupMentionRepo  domainrepository.MessageGroupMentionRepository
	linkRepo          domainrepository.MessageLinkRepository
	attachmentRepo    domainrepository.AttachmentRepository
	userGroupRepo     domainrepository.UserGroupRepository
	messageAssembler  *message.MessageOutputAssembler
	channelAccessSvc  domainservice.ChannelAccessService
}

func NewBookmarkInteractor(
	bookmarkRepo domainrepository.BookmarkRepository,
	messageRepo domainrepository.MessageRepository,
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
	mentionRepo domainrepository.MessageUserMentionRepository,
	groupMentionRepo domainrepository.MessageGroupMentionRepository,
	linkRepo domainrepository.MessageLinkRepository,
	attachmentRepo domainrepository.AttachmentRepository,
	userGroupRepo domainrepository.UserGroupRepository,
	channelAccessSvc domainservice.ChannelAccessService,
) BookmarkUseCase {
	return &bookmarkInteractor{
		bookmarkRepo:      bookmarkRepo,
		messageRepo:       messageRepo,
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		workspaceRepo:     workspaceRepo,
		userRepo:          userRepo,
		mentionRepo:       mentionRepo,
		groupMentionRepo:  groupMentionRepo,
		linkRepo:          linkRepo,
		attachmentRepo:    attachmentRepo,
		userGroupRepo:     userGroupRepo,
		messageAssembler:  message.NewMessageOutputAssembler(),
		channelAccessSvc:  channelAccessSvc,
	}
}

func (i *bookmarkInteractor) AddBookmark(ctx context.Context, input AddBookmarkInput) error {
	// メッセージの存在確認とアクセス権限チェック
	message, err := i.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return fmt.Errorf("failed to fetch message: %w", err)
	}
	if message == nil {
		return ErrMessageNotFound
	}

	// チャンネルへのアクセス権限チェック
	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, message.ChannelID, input.UserID); err != nil {
		return err
	}

	// 既にブックマーク済みかチェック
	isBookmarked, err := i.bookmarkRepo.IsBookmarked(ctx, input.UserID, input.MessageID)
	if err != nil {
		return fmt.Errorf("failed to check bookmark status: %w", err)
	}
	if isBookmarked {
		return ErrBookmarkExists
	}

	// ブックマークを追加
	bookmark := &entity.MessageBookmark{
		UserID:    input.UserID,
		MessageID: input.MessageID,
		CreatedAt: time.Now(),
	}

	if err := i.bookmarkRepo.AddBookmark(ctx, bookmark); err != nil {
		return fmt.Errorf("failed to add bookmark: %w", err)
	}

	return nil
}

func (i *bookmarkInteractor) RemoveBookmark(ctx context.Context, input RemoveBookmarkInput) error {
	// メッセージの存在確認
	message, err := i.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return fmt.Errorf("failed to fetch message: %w", err)
	}
	if message == nil {
		return ErrMessageNotFound
	}

	// チャンネルへのアクセス権限チェック
	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, message.ChannelID, input.UserID); err != nil {
		return err
	}

	// ブックマークを削除
	if err := i.bookmarkRepo.RemoveBookmark(ctx, input.UserID, input.MessageID); err != nil {
		return fmt.Errorf("failed to remove bookmark: %w", err)
	}

	return nil
}

func (i *bookmarkInteractor) ListBookmarks(ctx context.Context, userID string) (*ListBookmarksOutput, error) {
	// ブックマーク一覧を取得
	bookmarks, err := i.bookmarkRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bookmarks: %w", err)
	}

	if len(bookmarks) == 0 {
		return &ListBookmarksOutput{Bookmarks: []BookmarkWithMessageOutput{}}, nil
	}

	// メッセージIDを収集
	messageIDs := make([]string, 0, len(bookmarks))
	for _, bookmark := range bookmarks {
		if bookmark.Message != nil {
			messageIDs = append(messageIDs, bookmark.Message.ID)
		}
	}

	// 関連データを一括取得
	var userMentions []*entity.MessageUserMention
	if i.mentionRepo != nil && len(messageIDs) > 0 {
		userMentions, _ = i.mentionRepo.FindByMessageIDs(ctx, messageIDs)
	}

	var groupMentions []*entity.MessageGroupMention
	if i.groupMentionRepo != nil && len(messageIDs) > 0 {
		groupMentions, _ = i.groupMentionRepo.FindByMessageIDs(ctx, messageIDs)
	}

	var links []*entity.MessageLink
	if i.linkRepo != nil && len(messageIDs) > 0 {
		links, _ = i.linkRepo.FindByMessageIDs(ctx, messageIDs)
	}

	reactions := make(map[string][]*entity.MessageReaction)
	if len(messageIDs) > 0 {
		if result, err := i.messageRepo.FindReactionsByMessageIDs(ctx, messageIDs); err == nil {
			reactions = result
		}
	}

	attachments := make(map[string][]*entity.Attachment)
	if i.attachmentRepo != nil && len(messageIDs) > 0 {
		if result, err := i.attachmentRepo.FindByMessageIDs(ctx, messageIDs); err == nil {
			attachments = result
		}
	}

	// ユーザーIDを収集
	userIDSet := make(map[string]bool)
	userIDList := make([]string, 0)
	for _, bookmark := range bookmarks {
		if bookmark.Message != nil && !userIDSet[bookmark.Message.UserID] {
			userIDList = append(userIDList, bookmark.Message.UserID)
			userIDSet[bookmark.Message.UserID] = true
		}
	}
	for _, reactionList := range reactions {
		for _, reaction := range reactionList {
			if !userIDSet[reaction.UserID] {
				userIDList = append(userIDList, reaction.UserID)
				userIDSet[reaction.UserID] = true
			}
		}
	}

	// ユーザー情報を一括取得
	userMap := make(map[string]*entity.User)
	if i.userRepo != nil && len(userIDList) > 0 {
		users, _ := i.userRepo.FindByIDs(ctx, userIDList)
		for _, user := range users {
			userMap[user.ID] = user
		}
	}

	// グループIDを収集
	groupIDSet := make(map[string]bool)
	groupIDList := make([]string, 0)
	for _, mention := range groupMentions {
		if !groupIDSet[mention.GroupID] {
			groupIDList = append(groupIDList, mention.GroupID)
			groupIDSet[mention.GroupID] = true
		}
	}

	// グループ情報を一括取得
	groups := make(map[string]*entity.UserGroup)
	if i.userGroupRepo != nil && len(groupIDList) > 0 {
		groupList, err := i.userGroupRepo.FindByIDs(ctx, groupIDList)
		if err == nil {
			for _, group := range groupList {
				groups[group.ID] = group
			}
		}
	}

	// メッセージIDごとにデータをグループ化
	userMentionsByMessage := make(map[string][]*entity.MessageUserMention)
	for _, mention := range userMentions {
		userMentionsByMessage[mention.MessageID] = append(userMentionsByMessage[mention.MessageID], mention)
	}

	groupMentionsByMessage := make(map[string][]*entity.MessageGroupMention)
	for _, mention := range groupMentions {
		groupMentionsByMessage[mention.MessageID] = append(groupMentionsByMessage[mention.MessageID], mention)
	}

	linksByMessage := make(map[string][]*entity.MessageLink)
	for _, link := range links {
		linksByMessage[link.MessageID] = append(linksByMessage[link.MessageID], link)
	}

	// BookmarkWithMessageOutputに変換
	outputs := make([]BookmarkWithMessageOutput, 0, len(bookmarks))
	for _, bookmark := range bookmarks {
		if bookmark.Message == nil {
			continue
		}

		messageOutput := i.messageAssembler.AssembleMessageOutput(
			bookmark.Message,
			userMap[bookmark.Message.UserID],
			userMentionsByMessage[bookmark.Message.ID],
			groupMentionsByMessage[bookmark.Message.ID],
			linksByMessage[bookmark.Message.ID],
			reactions[bookmark.Message.ID],
			attachments[bookmark.Message.ID],
			groups,
			userMap,
		)

		outputs = append(outputs, BookmarkWithMessageOutput{
			UserID:    bookmark.UserID,
			Message:   messageOutput,
			CreatedAt: bookmark.CreatedAt,
		})
	}

	return &ListBookmarksOutput{Bookmarks: outputs}, nil
}

func (i *bookmarkInteractor) IsBookmarked(ctx context.Context, userID, messageID string) (bool, error) {
	return i.bookmarkRepo.IsBookmarked(ctx, userID, messageID)
}

// ensureChannelAccess は ChannelAccessService に委譲済み
