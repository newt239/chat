package repository

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/newt239/chat/ent/message"
	"github.com/newt239/chat/ent/messagereaction"
	"github.com/newt239/chat/ent/session"
	"github.com/newt239/chat/ent/workspace"
	"github.com/newt239/chat/ent/workspacemember"
	"github.com/newt239/chat/internal/domain/entity"
)

func TestUserRepositoryPreferencesAndLinks(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	repo := NewUserRepository(client)
	ctx := context.Background()

	u, err := repo.FindByID(ctx, f.alice.ID.String())
	if err != nil {
		t.Fatalf("取得に失敗しました: %v", err)
	}
	if u.Preferences != entity.DefaultPreferences() || len(u.Links) != 0 {
		t.Fatalf("設定を保存していなければ既定値を返すことを期待しましたが %+v %v でした", u.Preferences, u.Links)
	}

	for _, links := range [][]string{{"https://a.example", "https://b.example"}, {"https://c.example"}} {
		u.Preferences.Locale = "en"
		u.Links = links
		if err := repo.Update(ctx, u); err != nil {
			t.Fatalf("更新に失敗しました: %v", err)
		}
		got, err := repo.FindByID(ctx, u.ID)
		if err != nil {
			t.Fatalf("取得に失敗しました: %v", err)
		}
		if got.Preferences.Locale != "en" || !reflect.DeepEqual(got.Links, links) {
			t.Errorf("設定とリンクが保存されていません: %+v %v", got.Preferences, got.Links)
		}
	}
}

func TestUserRepositoryDeleteKeepsPostsAndAnonymizes(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	ctx := context.Background()
	alice, general, thread := f.alice, f.channels["general"], f.messages["mention"]

	client.Session.Create().SetUser(alice).SetRefreshTokenHash("h-" + alice.ID.String()).SetExpiresAt(time.Now().Add(time.Hour)).SaveX(ctx)
	client.PushToken.Create().SetUser(alice).SetToken("t-" + alice.ID.String()).SetPlatform("web").SaveX(ctx)
	client.ChannelReadState.Create().SetChannel(general).SetUser(alice).SaveX(ctx)
	client.ThreadReadState.Create().SetUser(alice).SetThread(thread).SaveX(ctx)
	client.UserThreadFollow.Create().SetUser(alice).SetThread(thread).SaveX(ctx)
	client.ChannelStar.Create().SetUser(alice).SetChannel(general).SaveX(ctx)
	client.ChannelMute.Create().SetUser(alice).SetChannel(general).SaveX(ctx)
	client.Draft.Create().SetUser(alice).SetChannel(general).SetBody("下書き").SaveX(ctx)
	client.ScheduledMessage.Create().SetUser(alice).SetChannel(general).SetBody("予約").SetScheduledAt(time.Now().Add(time.Hour)).SaveX(ctx)
	client.Reminder.Create().SetWorkspaceID(f.workspaceID).SetCreatorID(alice.ID).SetText("remind").SetRemindAt(time.Now().Add(time.Hour)).SaveX(ctx)
	client.MessageBookmark.Create().SetUser(alice).SetMessage(thread).SaveX(ctx)
	client.UserNote.Create().SetOwnerID(alice.ID).SetTargetID(f.bob.ID).SetNickname("ボブ").SaveX(ctx)
	category := client.ChannelCategory.Create().SetUser(alice).SetWorkspaceID(f.workspaceID).SetName("cat").SaveX(ctx)
	client.ChannelCategoryItem.Create().SetCategory(category).SetUser(alice).SetChannel(general).SaveX(ctx)
	client.MessageReaction.Create().SetMessageID(thread.ID).SetUser(alice).SetEmoji("👍").SaveX(ctx)
	repo := NewUserRepository(client)
	u, err := repo.FindByID(ctx, alice.ID.String())
	if err != nil {
		t.Fatalf("取得に失敗しました: %v", err)
	}
	googleSub, bio := "sub-"+alice.ID.String(), "bio"
	u.GoogleSub, u.Bio, u.Links = &googleSub, &bio, []string{"https://a.example"}
	if err := repo.Update(ctx, u); err != nil {
		t.Fatalf("更新に失敗しました: %v", err)
	}

	if err := repo.Delete(ctx, alice.ID.String()); err != nil {
		t.Fatalf("退会が外部キーなどで失敗しました: %v", err)
	}

	got, err := repo.FindByID(ctx, alice.ID.String())
	if err != nil || got == nil {
		t.Fatalf("ユーザーの行が残っていません: %v", err)
	}
	if got.DeletedAt == nil || got.DisplayName != entity.DeletedUserDisplayName || got.Email == alice.Email ||
		got.GoogleSub != nil || got.Bio != nil || got.AvatarURL != nil || len(got.Links) != 0 || got.PasswordHash != entity.UnusablePasswordHash {
		t.Errorf("匿名化されていません: %+v", got)
	}
	if n := client.WorkspaceMember.Query().Where(workspacemember.UserID(alice.ID)).CountX(ctx); n != 0 {
		t.Errorf("ワークスペースのメンバー情報が残っています: %d", n)
	}
	if n := client.Session.Query().Where(session.UserID(alice.ID)).CountX(ctx); n != 0 {
		t.Errorf("セッションが残っています: %d", n)
	}
	if !client.Message.Query().Where(message.ID(f.messages["reply"].ID)).ExistX(ctx) || !client.Workspace.Query().Where(workspace.ID(f.workspaceID)).ExistX(ctx) ||
		!client.MessageReaction.Query().Where(messagereaction.UserID(alice.ID)).ExistX(ctx) {
		t.Error("投稿・ワークスペース・リアクションは残すはず")
	}
}
