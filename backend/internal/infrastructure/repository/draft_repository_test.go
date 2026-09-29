package repository

import (
	"context"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

func TestDraftRepositoryUpsertKeepsOnePerTarget(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	repo := NewDraftRepository(client)
	ctx := context.Background()
	userID, channelID := f.alice.ID.String(), f.channels["general"].ID.String()
	parentID := f.messages["mention"].ID.String()

	save := func(parent *string, body string) *entity.Draft {
		t.Helper()
		d := &entity.Draft{UserID: userID, ChannelID: channelID, ParentID: parent, Body: body}
		if err := repo.Upsert(ctx, d); err != nil {
			t.Fatalf("保存に失敗しました: %v", err)
		}
		return d
	}
	first := save(nil, "一回目")
	second := save(nil, "二回目")
	save(&parentID, "スレッド")
	save(&parentID, "スレッド二回目")

	if first.ID != second.ID {
		t.Errorf("同じチャンネルの下書きが別の行になっています: %s %s", first.ID, second.ID)
	}
	drafts, err := repo.FindByWorkspace(ctx, userID, f.workspaceID)
	if err != nil || len(drafts) != 2 {
		t.Fatalf("チャンネルとスレッドの 2 件になっていません: %d %v", len(drafts), err)
	}
	thread, err := repo.Find(ctx, domainrepository.DraftTarget{UserID: userID, ChannelID: channelID, ParentID: &parentID})
	if err != nil || thread == nil || thread.Body != "スレッド二回目" {
		t.Fatalf("スレッドの下書きが上書きされていません: %+v %v", thread, err)
	}

	if err := repo.Delete(ctx, domainrepository.DraftTarget{UserID: userID, ChannelID: channelID}); err != nil {
		t.Fatalf("削除に失敗しました: %v", err)
	}
	channelDraft, _ := repo.Find(ctx, domainrepository.DraftTarget{UserID: userID, ChannelID: channelID})
	thread, _ = repo.Find(ctx, domainrepository.DraftTarget{UserID: userID, ChannelID: channelID, ParentID: &parentID})
	if channelDraft != nil || thread == nil {
		t.Errorf("チャンネルの下書きだけを削除できていません: channel=%+v thread=%+v", channelDraft, thread)
	}
}
