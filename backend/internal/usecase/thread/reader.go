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
	if err := r.ensureThreadAccess(ctx, input.ThreadID, input.UserID); err != nil {
		return err
	}

	if err := r.threadRepo.UpsertReadState(ctx, input.UserID, input.ThreadID, time.Now()); err != nil {
		return fmt.Errorf("failed to mark thread as read: %w", err)
	}
	return nil
}

// FollowThread はスレッドをフォローします
func (r *ThreadReader) FollowThread(ctx context.Context, input FollowThreadInput) error {
	if err := r.ensureThreadAccess(ctx, input.ThreadID, input.UserID); err != nil {
		return err
	}

	if err := r.threadRepo.FollowThread(ctx, input.UserID, input.ThreadID); err != nil {
		return fmt.Errorf("failed to follow thread: %w", err)
	}
	return nil
}

// UnfollowThread はスレッドのフォローを解除します
func (r *ThreadReader) UnfollowThread(ctx context.Context, input FollowThreadInput) error {
	if err := r.ensureThreadAccess(ctx, input.ThreadID, input.UserID); err != nil {
		return err
	}

	if err := r.threadRepo.UnfollowThread(ctx, input.UserID, input.ThreadID); err != nil {
		return fmt.Errorf("failed to unfollow thread: %w", err)
	}
	return nil
}

// ensureThreadAccess はスレッドの存在とチャンネルへのアクセス権を確認します
func (r *ThreadReader) ensureThreadAccess(ctx context.Context, threadID, userID string) error {
	thread, err := r.messageRepo.FindByID(ctx, threadID)
	if err != nil {
		return fmt.Errorf("failed to fetch thread: %w", err)
	}
	if thread == nil {
		return domainerrors.ErrMessageNotFound
	}

	_, err = r.channelAccessSvc.EnsureChannelAccess(ctx, thread.ChannelID, userID)
	return err
}
