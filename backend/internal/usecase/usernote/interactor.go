package usernote

import (
	"context"
	"fmt"
	"strings"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type Interactor struct {
	userNoteRepo domainrepository.UserNoteRepository
	userRepo     domainrepository.UserRepository
}

func New(userNoteRepo domainrepository.UserNoteRepository, userRepo domainrepository.UserRepository) *Interactor {
	return &Interactor{userNoteRepo: userNoteRepo, userRepo: userRepo}
}

// Get は未設定の場合 nil を返します
func (i *Interactor) Get(ctx context.Context, ownerID, targetID string) (*entity.UserNote, error) {
	note, err := i.userNoteRepo.Find(ctx, ownerID, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user note: %w", err)
	}
	return note, nil
}

// Update はニックネームとメモを保存します。空白だけの値は未設定として扱い、両方とも空なら削除して nil を返します
func (i *Interactor) Update(ctx context.Context, note entity.UserNote) (*entity.UserNote, error) {
	target, err := i.userRepo.FindByID(ctx, note.TargetID)
	if err != nil {
		return nil, fmt.Errorf("failed to load target user: %w", err)
	}
	if target == nil {
		return nil, domerr.ErrUserNotFound
	}
	note.Nickname, note.Memo = nonEmpty(note.Nickname), nonEmpty(note.Memo)
	if note.Nickname == nil && note.Memo == nil {
		if err := i.userNoteRepo.Delete(ctx, note.OwnerID, note.TargetID); err != nil {
			return nil, fmt.Errorf("failed to delete user note: %w", err)
		}
		return nil, nil
	}
	if err := i.userNoteRepo.Upsert(ctx, &note); err != nil {
		return nil, fmt.Errorf("failed to save user note: %w", err)
	}
	return &note, nil
}

func nonEmpty(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
