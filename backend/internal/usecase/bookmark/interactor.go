package bookmark

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/example/chat/internal/domain/entity"
	domainrepository "github.com/example/chat/internal/domain/repository"
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
	bookmarkRepo  domainrepository.BookmarkRepository
	messageRepo   domainrepository.MessageRepository
	channelRepo   domainrepository.ChannelRepository
	workspaceRepo domainrepository.WorkspaceRepository
}

func NewBookmarkInteractor(
	bookmarkRepo domainrepository.BookmarkRepository,
	messageRepo domainrepository.MessageRepository,
	channelRepo domainrepository.ChannelRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
) BookmarkUseCase {
	return &bookmarkInteractor{
		bookmarkRepo:  bookmarkRepo,
		messageRepo:   messageRepo,
		channelRepo:   channelRepo,
		workspaceRepo: workspaceRepo,
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

	// BookmarkOutputに変換
	outputs := make([]BookmarkOutput, 0, len(bookmarks))
	for _, bookmark := range bookmarks {
		outputs = append(outputs, BookmarkOutput{
			UserID:    bookmark.UserID,
			MessageID: bookmark.MessageID,
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
		isMember, err := i.channelRepo.IsMember(ctx, ch.ID, userID)
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
