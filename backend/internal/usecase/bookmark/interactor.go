package bookmark

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
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
	bookmarkRepo     domainrepository.BookmarkRepository
	messageRepo      domainrepository.MessageRepository
	outputBuilder    *message.MessageOutputBuilder
	channelAccessSvc service.ChannelAccessService
}

func NewBookmarkInteractor(
	bookmarkRepo domainrepository.BookmarkRepository,
	messageRepo domainrepository.MessageRepository,
	outputBuilder *message.MessageOutputBuilder,
	channelAccessSvc service.ChannelAccessService,
) BookmarkUseCase {
	return &bookmarkInteractor{
		bookmarkRepo:     bookmarkRepo,
		messageRepo:      messageRepo,
		outputBuilder:    outputBuilder,
		channelAccessSvc: channelAccessSvc,
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

	bookmarked := make([]*entity.MessageBookmark, 0, len(bookmarks))
	messages := make([]*entity.Message, 0, len(bookmarks))
	for _, bookmark := range bookmarks {
		if bookmark.Message != nil {
			bookmarked = append(bookmarked, bookmark)
			messages = append(messages, bookmark.Message)
		}
	}
	messageOutputs, err := i.outputBuilder.Build(ctx, userID, messages)
	if err != nil {
		return nil, err
	}

	outputs := make([]BookmarkWithMessageOutput, 0, len(messageOutputs))
	for idx, messageOutput := range messageOutputs {
		outputs = append(outputs, BookmarkWithMessageOutput{
			UserID:    bookmarked[idx].UserID,
			Message:   messageOutput,
			CreatedAt: bookmarked[idx].CreatedAt,
		})
	}

	return &ListBookmarksOutput{Bookmarks: outputs}, nil
}

func (i *bookmarkInteractor) IsBookmarked(ctx context.Context, userID, messageID string) (bool, error) {
	return i.bookmarkRepo.IsBookmarked(ctx, userID, messageID)
}

// ensureChannelAccess は ChannelAccessService に委譲済み
