package usernote

import (
	"context"
	"fmt"
	"strings"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type UseCase interface {
	Get(ctx context.Context, input GetInput) (*Output, error)
	Update(ctx context.Context, input UpdateInput) (*Output, error)
}

type interactor struct {
	userNoteRepo domainrepository.UserNoteRepository
	userRepo     domainrepository.UserRepository
}

func NewInteractor(userNoteRepo domainrepository.UserNoteRepository, userRepo domainrepository.UserRepository) UseCase {
	return &interactor{userNoteRepo: userNoteRepo, userRepo: userRepo}
}

// Get は未設定の場合 nil を返します
func (i *interactor) Get(ctx context.Context, input GetInput) (*Output, error) {
	note, err := i.userNoteRepo.Find(ctx, input.OwnerID, input.TargetID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user note: %w", err)
	}
	if note == nil {
		return nil, nil
	}
	return toOutput(note), nil
}

// Update はニックネームとメモを保存します。両方とも空の場合は削除して nil を返します
func (i *interactor) Update(ctx context.Context, input UpdateInput) (*Output, error) {
	target, err := i.userRepo.FindByID(ctx, input.TargetID)
	if err != nil {
		return nil, fmt.Errorf("failed to load target user: %w", err)
	}
	if target == nil {
		return nil, entity.ErrUserNotFound
	}

	note := &entity.UserNote{
		OwnerID:  input.OwnerID,
		TargetID: input.TargetID,
		Nickname: nonEmpty(input.Nickname),
		Memo:     nonEmpty(input.Memo),
	}
	if note.Nickname == nil && note.Memo == nil {
		if err := i.userNoteRepo.Delete(ctx, input.OwnerID, input.TargetID); err != nil {
			return nil, fmt.Errorf("failed to delete user note: %w", err)
		}
		return nil, nil
	}

	if err := i.userNoteRepo.Upsert(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to save user note: %w", err)
	}
	return toOutput(note), nil
}

func nonEmpty(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func toOutput(note *entity.UserNote) *Output {
	return &Output{
		TargetID:  note.TargetID,
		Nickname:  note.Nickname,
		Memo:      note.Memo,
		UpdatedAt: note.UpdatedAt,
	}
}
