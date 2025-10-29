package thread

import (
	"context"
	"fmt"
	"time"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type ThreadReader struct {
	threadRepo domainrepository.ThreadRepository
}

func NewThreadReader(
	threadRepo domainrepository.ThreadRepository,
) *ThreadReader {
	return &ThreadReader{
		threadRepo: threadRepo,
	}
}

func (r *ThreadReader) MarkThreadRead(ctx context.Context, input MarkThreadReadInput) error {
	now := time.Now()
	err := r.threadRepo.UpsertReadState(ctx, input.UserID, input.ThreadID, now)
	if err != nil {
		return fmt.Errorf("failed to mark thread as read: %w", err)
	}
	return nil
}
