package bookmark

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/usecase/message"
)

var (
	ErrMessageNotFound = errors.New("message not found")
	ErrUnauthorized    = errors.New("unauthorized to perform this action")
	ErrBookmarkExists  = errors.New("bookmark already exists")
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
	if err := i.ensureChannelAccess(ctx, message.ChannelID, input.UserID); err != nil {
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
	if err := i.ensureChannelAccess(ctx, message.ChannelID, input.UserID); err != nil {
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
	userMentions, _ := i.mentionRepo.FindByMessageIDs(ctx, messageIDs)
	groupMentions, _ := i.groupMentionRepo.FindByMessageIDs(ctx, messageIDs)
	links, _ := i.linkRepo.FindByMessageIDs(ctx, messageIDs)
	reactions, _ := i.messageRepo.FindReactionsByMessageIDs(ctx, messageIDs)
	attachments, _ := i.attachmentRepo.FindByMessageIDs(ctx, messageIDs)

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
	users, _ := i.userRepo.FindByIDs(ctx, userIDList)
	userMap := make(map[string]*entity.User)
	for _, user := range users {
		userMap[user.ID] = user
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
	if len(groupIDList) > 0 {
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

func (i *bookmarkInteractor) ensureChannelAccess(ctx context.Context, channelID, userID string) error {
	ch, err := i.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to load channel: %w", err)
	}
	if ch == nil {
		return errors.New("channel not found")
	}

	// プライベートチャンネルの場合
	if ch.IsPrivate {
		isMember, err := i.channelMemberRepo.IsMember(ctx, ch.ID, userID)
		if err != nil {
			return fmt.Errorf("failed to verify channel membership: %w", err)
		}
		if !isMember {
			return ErrUnauthorized
		}
		return nil
	}

	// パブリックチャンネルの場合はワークスペースメンバーかチェック
	member, err := i.workspaceRepo.FindMember(ctx, ch.WorkspaceID, userID)
	if err != nil {
		return fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return ErrUnauthorized
	}

	return nil
}
