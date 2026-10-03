package repository

import (
	"context"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
)

func TestBookmarkFindByUserIDExcludesDeletedAndOtherWorkspaces(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	ctx := context.Background()

	other := client.Workspace.Create().SetID("o" + f.workspaceID).SetName("other").SetCreatedBy(f.alice).SaveX(ctx)
	ch := client.Channel.Create().SetName("general").SetChannelType(string(entity.ChannelTypePublic)).SetWorkspace(other).SetCreatedBy(f.alice).SaveX(ctx)
	elsewhere := client.Message.Create().SetChannel(ch).SetUser(f.alice).SetBody("別のワークスペース").SaveX(ctx)
	for _, m := range []string{"mention", "deleted"} {
		client.MessageBookmark.Create().SetUser(f.alice).SetMessage(f.messages[m]).SaveX(ctx)
	}
	client.MessageBookmark.Create().SetUser(f.alice).SetMessage(elsewhere).SaveX(ctx)

	got, err := NewBookmarkRepository(client).FindByUserID(ctx, f.alice.ID.String(), f.workspaceID)
	if err != nil {
		t.Fatalf("取得に失敗しました: %v", err)
	}
	if len(got) != 1 || got[0].MessageID != f.messages["mention"].ID.String() {
		t.Fatalf("削除済みと別のワークスペースのブックマークを除くはず: %+v", got)
	}
}
