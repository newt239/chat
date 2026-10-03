package bookmark

import (
	"context"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/message"
)

type BookmarkUseCase interface {
	AddBookmark(ctx context.Context, input AddBookmarkInput) error
	RemoveBookmark(ctx context.Context, input RemoveBookmarkInput) error
	ListBookmarks(ctx context.Context, userID string) (*ListBookmarksOutput, error)
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

// ensureAccess はメッセージがあり、そのチャンネルを閲覧できることを確かめます
func (i *bookmarkInteractor) ensureAccess(ctx context.Context, messageID, userID string) error {
	msg, err := i.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to fetch message: %w", err)
	}
	if msg == nil {
		return domerr.ErrMessageNotFound
	}
	_, err = i.channelAccessSvc.EnsureChannelAccess(ctx, msg.ChannelID, userID)
	return err
}

// AddBookmark は既にブックマークしていれば ErrBookmarkExists を返します
func (i *bookmarkInteractor) AddBookmark(ctx context.Context, input AddBookmarkInput) error {
	if err := i.ensureAccess(ctx, input.MessageID, input.UserID); err != nil {
		return err
	}
	return i.bookmarkRepo.AddBookmark(ctx, &entity.MessageBookmark{UserID: input.UserID, MessageID: input.MessageID})
}

func (i *bookmarkInteractor) RemoveBookmark(ctx context.Context, input RemoveBookmarkInput) error {
	if err := i.ensureAccess(ctx, input.MessageID, input.UserID); err != nil {
		return err
	}
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
