package bookmark

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/message"
)

type Output struct {
	Message   message.MessageOutput
	CreatedAt time.Time
}

type Interactor struct {
	bookmarkRepo     domainrepository.BookmarkRepository
	outputBuilder    *message.MessageOutputBuilder
	channelAccessSvc service.ChannelAccessService
}

func New(
	bookmarkRepo domainrepository.BookmarkRepository,
	outputBuilder *message.MessageOutputBuilder,
	channelAccessSvc service.ChannelAccessService,
) *Interactor {
	return &Interactor{bookmarkRepo: bookmarkRepo, outputBuilder: outputBuilder, channelAccessSvc: channelAccessSvc}
}

// AddBookmark は既にブックマークしていれば ErrBookmarkExists を返します
func (i *Interactor) AddBookmark(ctx context.Context, input message.MessageInput) error {
	if _, _, err := i.channelAccessSvc.EnsureMessageAccess(ctx, input.MessageID, input.UserID); err != nil {
		return err
	}
	return i.bookmarkRepo.AddBookmark(ctx, &entity.MessageBookmark{UserID: input.UserID, MessageID: input.MessageID})
}

func (i *Interactor) RemoveBookmark(ctx context.Context, input message.MessageInput) error {
	if _, _, err := i.channelAccessSvc.EnsureMessageAccess(ctx, input.MessageID, input.UserID); err != nil {
		return err
	}
	if err := i.bookmarkRepo.RemoveBookmark(ctx, input.UserID, input.MessageID); err != nil {
		return fmt.Errorf("failed to remove bookmark: %w", err)
	}
	return nil
}

// ListBookmarks はワークスペース内で今も閲覧できるメッセージのブックマークだけを返します
func (i *Interactor) ListBookmarks(ctx context.Context, userID, workspaceID string) ([]Output, error) {
	bookmarks, err := i.bookmarkRepo.FindByUserID(ctx, userID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bookmarks: %w", err)
	}
	channelIDs := make([]string, len(bookmarks))
	for idx, bookmark := range bookmarks {
		channelIDs[idx] = bookmark.Message.ChannelID
	}
	channels, err := i.channelAccessSvc.AccessibleChannelsByIDs(ctx, channelIDs, userID)
	if err != nil {
		return nil, err
	}
	bookmarks = slices.DeleteFunc(bookmarks, func(b *entity.MessageBookmark) bool { return channels[b.Message.ChannelID] == nil })
	messages := make([]*entity.Message, len(bookmarks))
	for idx, bookmark := range bookmarks {
		messages[idx] = bookmark.Message
	}
	messageOutputs, err := i.outputBuilder.Build(ctx, userID, messages)
	if err != nil {
		return nil, err
	}
	outputs := make([]Output, len(bookmarks))
	for idx, bookmark := range bookmarks {
		outputs[idx] = Output{Message: messageOutputs[idx], CreatedAt: bookmark.CreatedAt}
	}
	return outputs, nil
}
