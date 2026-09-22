package thread

import (
	"context"
	"fmt"
	"time"

	domainerrors "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
)

type ThreadReader struct {
	threadRepo       domainrepository.ThreadRepository
	messageRepo      domainrepository.MessageRepository
	channelAccessSvc domainservice.ChannelAccessService
}

func NewThreadReader(
	threadRepo domainrepository.ThreadRepository,
	messageRepo domainrepository.MessageRepository,
	channelAccessSvc domainservice.ChannelAccessService,
) *ThreadReader {
	return &ThreadReader{
		threadRepo:       threadRepo,
		messageRepo:      messageRepo,
		channelAccessSvc: channelAccessSvc,
	}
}

func (r *ThreadReader) MarkThreadRead(ctx context.Context, input MarkThreadReadInput) error {
	thread, err := r.messageRepo.FindByID(ctx, input.ThreadID)
	if err != nil {
		return fmt.Errorf("failed to fetch thread: %w", err)
	}
	if thread == nil {
		return domainerrors.ErrMessageNotFound
	}

	if _, err := r.channelAccessSvc.EnsureChannelAccess(ctx, thread.ChannelID, input.UserID); err != nil {
		return err
	}

	if err := r.threadRepo.UpsertReadState(ctx, input.UserID, input.ThreadID, time.Now()); err != nil {
		return fmt.Errorf("failed to mark thread as read: %w", err)
	}
	return nil
}
