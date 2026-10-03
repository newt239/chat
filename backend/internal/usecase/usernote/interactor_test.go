package usernote

import (
	"context"
	"errors"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type fakeNoteRepo struct {
	domainrepository.UserNoteRepository
	note    *entity.UserNote
	deleted bool
}

func (r *fakeNoteRepo) Find(_ context.Context, _ string, _ string) (*entity.UserNote, error) {
	return r.note, nil
}

func (r *fakeNoteRepo) Upsert(_ context.Context, note *entity.UserNote) error {
	r.note = note
	return nil
}

func (r *fakeNoteRepo) Delete(_ context.Context, _ string, _ string) error {
	r.note = nil
	r.deleted = true
	return nil
}

type stubUserRepo struct {
	domainrepository.UserRepository
}

func (stubUserRepo) FindByID(_ context.Context, id string) (*entity.User, error) {
	if id == "missing" {
		return nil, nil
	}
	return &entity.User{ID: id}, nil
}

func TestUpdateUserNote(t *testing.T) {
	repo := &fakeNoteRepo{}
	uc := New(repo, stubUserRepo{})
	ctx := context.Background()

	out, err := uc.Update(ctx, entity.UserNote{OwnerID: "me", TargetID: "bob", Nickname: new("  ボブ  "), Memo: new("")})
	if err != nil {
		t.Fatalf("保存できません: %v", err)
	}
	if *out.Nickname != "ボブ" || out.Memo != nil {
		t.Fatalf("空白の除去や空文字の扱いが正しくありません: %+v", out)
	}

	out, err = uc.Update(ctx, entity.UserNote{OwnerID: "me", TargetID: "bob"})
	if err != nil || out != nil || !repo.deleted {
		t.Fatalf("両方とも空のときに削除されていません: out=%+v err=%v", out, err)
	}

	if _, err := uc.Update(ctx, entity.UserNote{OwnerID: "me", TargetID: "missing", Nickname: new("x")}); !errors.Is(err, domerr.ErrUserNotFound) {
		t.Fatalf("存在しないユーザーへのメモが拒否されていません: %v", err)
	}
}
