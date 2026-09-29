package draft

import (
	"context"
	"errors"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

type fakeDraftRepo struct {
	domainrepository.DraftRepository
	saved   *entity.Draft
	deleted bool
}

func (r *fakeDraftRepo) Upsert(_ context.Context, d *entity.Draft) error {
	d.ID = "d1"
	r.saved = d
	return nil
}

func (r *fakeDraftRepo) Delete(_ context.Context, _ domainrepository.DraftTarget) error {
	r.deleted = true
	return nil
}

type stubMessageRepo struct {
	domainrepository.MessageRepository
}

func (stubMessageRepo) FindByID(_ context.Context, id string) (*entity.Message, error) {
	if id == "other-channel" {
		return &entity.Message{ID: id, ChannelID: "ch2"}, nil
	}
	return &entity.Message{ID: id, ChannelID: "ch1"}, nil
}

type allowAccess struct {
	service.ChannelAccessService
}

func (allowAccess) EnsureChannelAccess(_ context.Context, channelID, _ string) (*entity.Channel, error) {
	return &entity.Channel{ID: channelID}, nil
}

func TestSaveDraft(t *testing.T) {
	ctx := context.Background()
	repo := &fakeDraftRepo{}
	uc := NewInteractor(repo, stubMessageRepo{}, allowAccess{})
	parent := "m1"

	d, err := uc.Save(ctx, SaveInput{Target: domainrepository.DraftTarget{UserID: "u1", ChannelID: "ch1", ParentID: &parent}, Body: "書きかけ"})
	if err != nil || d == nil || repo.saved.Body != "書きかけ" || *repo.saved.ParentID != "m1" {
		t.Fatalf("スレッドの下書きを保存できません: d=%+v err=%v", d, err)
	}

	d, err = uc.Save(ctx, SaveInput{Target: domainrepository.DraftTarget{UserID: "u1", ChannelID: "ch1"}, Body: " \n "})
	if err != nil || d != nil || !repo.deleted {
		t.Fatalf("空白だけの本文で削除されていません: d=%+v err=%v", d, err)
	}

	other := "other-channel"
	_, err = uc.Save(ctx, SaveInput{Target: domainrepository.DraftTarget{UserID: "u1", ChannelID: "ch1", ParentID: &other}, Body: "x"})
	if !errors.Is(err, ErrParentMessageNotFound) {
		t.Fatalf("別チャンネルのメッセージへの返信を拒否していません: %v", err)
	}
}
