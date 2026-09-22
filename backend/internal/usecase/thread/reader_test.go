package thread

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainerrors "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type stubMessageRepo struct {
	domainrepository.MessageRepository
	message *entity.Message
}

func (r *stubMessageRepo) FindByID(_ context.Context, _ string) (*entity.Message, error) {
	return r.message, nil
}

type stubChannelAccessService struct {
	err error
}

func (s *stubChannelAccessService) EnsureChannelAccess(_ context.Context, _ string, _ string) (*entity.Channel, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &entity.Channel{ID: "ch1"}, nil
}

type stubThreadRepo struct {
	domainrepository.ThreadRepository
	upsertCalls int
	followCalls int
}

func (r *stubThreadRepo) UpsertReadState(_ context.Context, _ string, _ string, _ time.Time) error {
	r.upsertCalls++
	return nil
}

func (r *stubThreadRepo) FollowThread(_ context.Context, _ string, _ string) error {
	r.followCalls++
	return nil
}

func TestMarkThreadReadRequiresExistingThread(t *testing.T) {
	threadRepo := &stubThreadRepo{}
	reader := NewThreadReader(threadRepo, &stubMessageRepo{}, &stubChannelAccessService{})

	err := reader.MarkThreadRead(context.Background(), MarkThreadReadInput{ThreadID: "t1", UserID: "u1"})

	if !errors.Is(err, domainerrors.ErrMessageNotFound) {
		t.Fatalf("存在しないスレッドが拒否されていません: %v", err)
	}
	if threadRepo.upsertCalls != 0 {
		t.Errorf("既読が更新されてしまいました: %d 回", threadRepo.upsertCalls)
	}
}

func TestMarkThreadReadRequiresChannelAccess(t *testing.T) {
	threadRepo := &stubThreadRepo{}
	reader := NewThreadReader(
		threadRepo,
		&stubMessageRepo{message: &entity.Message{ID: "t1", ChannelID: "ch1"}},
		&stubChannelAccessService{err: domainerrors.ErrUnauthorized},
	)

	err := reader.MarkThreadRead(context.Background(), MarkThreadReadInput{ThreadID: "t1", UserID: "u1"})

	if !errors.Is(err, domainerrors.ErrUnauthorized) {
		t.Fatalf("権限のないチャンネルのスレッドが拒否されていません: %v", err)
	}
	if threadRepo.upsertCalls != 0 {
		t.Errorf("既読が更新されてしまいました: %d 回", threadRepo.upsertCalls)
	}
}

func TestFollowThreadRequiresChannelAccess(t *testing.T) {
	threadRepo := &stubThreadRepo{}
	reader := NewThreadReader(
		threadRepo,
		&stubMessageRepo{message: &entity.Message{ID: "t1", ChannelID: "ch1"}},
		&stubChannelAccessService{err: domainerrors.ErrUnauthorized},
	)

	err := reader.FollowThread(context.Background(), FollowThreadInput{ThreadID: "t1", UserID: "u1"})

	if !errors.Is(err, domainerrors.ErrUnauthorized) {
		t.Fatalf("権限のないスレッドのフォローが拒否されていません: %v", err)
	}
	if threadRepo.followCalls != 0 {
		t.Errorf("フォローが登録されてしまいました: %d 回", threadRepo.followCalls)
	}
}

func TestMarkThreadReadSucceeds(t *testing.T) {
	threadRepo := &stubThreadRepo{}
	reader := NewThreadReader(
		threadRepo,
		&stubMessageRepo{message: &entity.Message{ID: "t1", ChannelID: "ch1"}},
		&stubChannelAccessService{},
	)

	if err := reader.MarkThreadRead(context.Background(), MarkThreadReadInput{ThreadID: "t1", UserID: "u1"}); err != nil {
		t.Fatalf("既読更新に失敗しました: %v", err)
	}
	if threadRepo.upsertCalls != 1 {
		t.Errorf("既読が更新されていません: %d 回", threadRepo.upsertCalls)
	}
}
