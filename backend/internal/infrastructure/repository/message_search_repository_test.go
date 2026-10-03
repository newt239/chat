package repository

import (
	"context"
	"os"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/migrate"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

// TEST_DATABASE_URL に空の PostgreSQL を指定したときだけ実行する
func openTestClient(t *testing.T) *ent.Client {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL が未設定のためスキップします")
	}
	client, err := ent.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("DB に接続できません: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	if err := client.Schema.Create(context.Background(), migrate.WithGlobalUniqueID(true), migrate.WithForeignKeys(true)); err != nil {
		t.Fatalf("マイグレーションに失敗しました: %v", err)
	}
	return client
}

type searchFixture struct {
	workspaceID string
	alice, bob  *ent.User
	channels    map[string]*ent.Channel
	group       *ent.UserGroup
	messages    map[string]*ent.Message
}

func newSearchFixture(t *testing.T, client *ent.Client) *searchFixture {
	t.Helper()
	ctx := context.Background()
	suffix := uuid.NewString()[:8]
	f := &searchFixture{workspaceID: "ws-" + suffix, channels: map[string]*ent.Channel{}, messages: map[string]*ent.Message{}}

	newUser := func(name string) *ent.User {
		return client.User.Create().SetEmail(name + suffix + "@example.com").SetPasswordHash("x").SetDisplayName(name).SaveX(ctx)
	}
	f.alice, f.bob = newUser("alice"), newUser("bob")
	client.Workspace.Create().SetID(f.workspaceID).SetName("ws").SetCreatedBy(f.alice).SaveX(ctx)
	for _, u := range []*ent.User{f.alice, f.bob} {
		client.WorkspaceMember.Create().SetWorkspaceID(f.workspaceID).SetUser(u).SetRole("member").SaveX(ctx)
	}

	newChannel := func(name string, private bool, members ...*ent.User) {
		channelType := entity.ChannelTypePublic
		if private {
			channelType = entity.ChannelTypePrivate
		}
		ch := client.Channel.Create().SetName(name).SetChannelType(string(channelType)).SetWorkspaceID(f.workspaceID).SetCreatedBy(f.bob).SaveX(ctx)
		for _, m := range members {
			client.ChannelMember.Create().SetChannel(ch).SetUser(m).SaveX(ctx)
		}
		f.channels[name] = ch
	}
	newChannel("general", false, f.alice, f.bob)
	newChannel("dev", false, f.alice, f.bob)
	newChannel("dev/web", false, f.bob)
	newChannel("secret", true, f.bob)

	f.group = client.UserGroup.Create().SetName("designers").SetWorkspaceID(f.workspaceID).SetCreatedBy(f.bob).SaveX(ctx)
	client.UserGroupMember.Create().SetGroup(f.group).SetUser(f.alice).SaveX(ctx)

	base := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	newMessage := func(key string, minute int, channel string, author *ent.User, body string) *ent.Message {
		m := client.Message.Create().SetChannel(f.channels[channel]).SetUser(author).SetBody(body).
			SetCreatedAt(base.Add(time.Duration(minute) * time.Minute)).SaveX(ctx)
		f.messages[key] = m
		return m
	}
	m1 := newMessage("mention", 0, "general", f.bob, "設計レビューをお願いします @alice")
	client.MessageUserMention.Create().SetMessage(m1).SetUser(f.alice).SaveX(ctx)
	newMessage("link", 1, "dev/web", f.bob, "Design review 資料 https://example.com")
	newMessage("secret", 2, "secret", f.bob, "設計 secret")
	broadcast := newMessage("broadcast", 3, "dev", f.bob, "<@channel> 設計の締め切りと設計書")
	f.messages["broadcast"] = broadcast.Update().SetMentionsChannel(true).SaveX(ctx)
	m5 := newMessage("group", 4, "general", f.bob, "group ping")
	client.MessageGroupMention.Create().SetMessage(m5).SetGroup(f.group).SaveX(ctx)
	client.MessageUserMention.Create().SetMessage(m5).SetUser(f.alice).SetViaGroupID(f.group.ID).SaveX(ctx)
	reply := client.Message.Create().SetChannel(f.channels["general"]).SetUser(f.alice).SetParent(m1).SetBody("了解 設計").
		SetCreatedAt(base.Add(5 * time.Minute)).SaveX(ctx)
	f.messages["reply"] = reply
	m7 := newMessage("image", 6, "dev", f.bob, "スクショ")
	client.Attachment.Create().SetFileName("screenshot.png").SetMimeType("image/png").SetSizeBytes(1).SetStorageKey("k").
		SetMessage(m7).SetUploader(f.bob).SetChannel(f.channels["dev"]).SaveX(ctx)
	client.MessagePin.Create().SetChannel(f.channels["dev"]).SetMessage(m7).SetPinnedBy(f.bob).SaveX(ctx)
	m8 := newMessage("self", 7, "general", f.alice, "@alice self")
	client.MessageUserMention.Create().SetMessage(m8).SetUser(f.alice).SaveX(ctx)
	f.messages["deleted"] = client.Message.Create().SetChannel(f.channels["general"]).SetUser(f.bob).SetBody("設計 deleted").
		SetCreatedAt(base.Add(8 * time.Minute)).SetDeletedAt(base.Add(9 * time.Minute)).SaveX(ctx)
	newMessage("not-broadcast", 9, "general", f.bob, "@channel は ID 記法ではないのでメンションではない")
	newMessage("reverse", 10, "general", f.bob, "review of the design")
	f.messages["location"] = client.Message.Create().SetChannel(f.channels["secret"]).SetUser(f.bob).SetBody("").
		SetLocationLatitude(35.68).SetLocationLongitude(139.76).SetCreatedAt(base.Add(11 * time.Minute)).SaveX(ctx)
	return f
}

func (f *searchFixture) keys(t *testing.T, messages []*entity.Message) []string {
	t.Helper()
	byID := map[string]string{}
	for key, m := range f.messages {
		byID[m.ID.String()] = key
	}
	keys := []string{}
	for _, m := range messages {
		keys = append(keys, byID[m.ID])
	}
	return keys
}

func TestFindSearchScope(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)

	scope, err := NewMessageRepository(client).FindSearchScope(context.Background(), f.workspaceID, f.alice.ID.String())
	if err != nil {
		t.Fatalf("取得に失敗しました: %v", err)
	}
	ids := func(names ...string) []string {
		result := []string{}
		for _, name := range names {
			result = append(result, f.channels[name].ID.String())
		}
		slices.Sort(result)
		return result
	}
	slices.Sort(scope.ViewableChannelIDs)
	slices.Sort(scope.JoinedChannelIDs)
	if !reflect.DeepEqual(scope.ViewableChannelIDs, ids("general", "dev", "dev/web")) {
		t.Errorf("閲覧できるチャンネルが期待と異なります: %v", scope.ViewableChannelIDs)
	}
	if !reflect.DeepEqual(scope.JoinedChannelIDs, ids("general", "dev")) {
		t.Errorf("参加しているチャンネルが期待と異なります: %v", scope.JoinedChannelIDs)
	}
}

func TestFindSearchDocuments(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	repo := NewMessageRepository(client)
	ids := []string{}
	for _, m := range f.messages {
		ids = append(ids, m.ID.String())
	}

	docs, err := repo.FindSearchDocuments(context.Background(), ids)
	if err != nil {
		t.Fatalf("取得に失敗しました: %v", err)
	}
	byKey := map[string]domainrepository.MessageSearchDocument{}
	for _, d := range docs {
		for key, m := range f.messages {
			if m.ID.String() == d.ID {
				byKey[key] = d
			}
		}
	}
	if _, ok := byKey["deleted"]; ok || len(docs) != len(f.messages)-1 {
		t.Fatalf("削除済みを除いた文書になっていません: %d 件", len(docs))
	}

	mention := byKey["mention"]
	if mention.WorkspaceID != f.workspaceID || mention.SenderID != f.bob.ID.String() || !mention.HasReplies ||
		!reflect.DeepEqual(mention.MentionedUserIDs, []string{f.alice.ID.String()}) {
		t.Errorf("メンションとスレッドが期待と異なります: %+v", mention)
	}
	if reply := byKey["reply"]; reply.ParentID == nil || *reply.ParentID != f.messages["mention"].ID.String() || reply.HasReplies {
		t.Errorf("返信の親が期待と異なります: %+v", reply)
	}
	image := byKey["image"]
	if !image.Pinned || !reflect.DeepEqual(image.AttachmentNames, []string{"screenshot.png"}) ||
		!reflect.DeepEqual(image.Has, []domainrepository.MessageContentKind{domainrepository.MessageContentImage}) {
		t.Errorf("添付とピン留めが期待と異なります: %+v", image)
	}
	if link := byKey["link"]; !reflect.DeepEqual(link.Has, []domainrepository.MessageContentKind{domainrepository.MessageContentLink}) {
		t.Errorf("リンクが期待と異なります: %+v", link.Has)
	}
	if location := byKey["location"]; !reflect.DeepEqual(location.Has, []domainrepository.MessageContentKind{domainrepository.MessageContentLocation}) {
		t.Errorf("位置情報が期待と異なります: %+v", location.Has)
	}
	if !byKey["broadcast"].MentionsChannel || byKey["not-broadcast"].MentionsChannel {
		t.Error("@channel の判定が期待と異なります")
	}
}

func TestFindSearchDocumentsAfterPagesByID(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	repo := NewMessageRepository(client)

	seen := map[string]bool{}
	after := ""
	for {
		docs, err := repo.FindSearchDocumentsAfter(context.Background(), after, 3)
		if err != nil {
			t.Fatalf("取得に失敗しました: %v", err)
		}
		if len(docs) == 0 {
			break
		}
		for _, d := range docs {
			if d.ID <= after || seen[d.ID] {
				t.Fatalf("ID 順に重複なく返っていません: after=%s id=%s", after, d.ID)
			}
			seen[d.ID] = true
		}
		after = docs[len(docs)-1].ID
	}
	for key, m := range f.messages {
		if seen[m.ID.String()] != (key != "deleted") {
			t.Errorf("%s の扱いが期待と異なります", key)
		}
	}
}

func TestFindMentions(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	repo := NewMessageRepository(client)
	input := domainrepository.FindMentionsInput{WorkspaceID: f.workspaceID, UserID: f.alice.ID.String(), Limit: 2}

	first, err := repo.FindMentions(context.Background(), input)
	if err != nil {
		t.Fatalf("取得に失敗しました: %v", err)
	}
	if got := f.keys(t, first); !reflect.DeepEqual(got, []string{"group", "broadcast"}) {
		t.Errorf("1 ページ目が期待と異なります: %v", got)
	}

	last := first[len(first)-1]
	input.Cursor = &domainrepository.MessageCursor{CreatedAt: last.CreatedAt, MessageID: last.ID}
	second, err := repo.FindMentions(context.Background(), input)
	if err != nil {
		t.Fatalf("取得に失敗しました: %v", err)
	}
	if got := f.keys(t, second); !reflect.DeepEqual(got, []string{"mention"}) {
		t.Errorf("2 ページ目が期待と異なります: %v", got)
	}
}

func TestFindParticipatingThreads(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	ctx := context.Background()
	repo := NewThreadRepository(client)

	// 古い投稿でも返信が新しいスレッドが先に来る
	old := client.Message.Create().SetChannel(f.channels["general"]).SetUser(f.bob).SetBody("古いスレッド").
		SetCreatedAt(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)).SaveX(ctx)
	for i, body := range []string{"a", "b", "c"} {
		client.Message.Create().SetChannel(f.channels["general"]).SetUser(f.bob).SetParent(old).SetBody(body).
			SetCreatedAt(time.Date(2026, 9, 2, 0, i, 0, 0, time.UTC)).SaveX(ctx)
	}
	client.UserThreadFollow.Create().SetUser(f.alice).SetThread(old).SaveX(ctx)

	input := domainrepository.FindParticipatingThreadsInput{WorkspaceID: f.workspaceID, UserID: f.alice.ID.String(), Limit: 1}
	first, err := repo.FindParticipatingThreads(ctx, input)
	if err != nil {
		t.Fatalf("取得に失敗しました: %v", err)
	}
	if len(first.Items) != 1 || first.Items[0].ThreadID != old.ID.String() {
		t.Fatalf("最新の返信があるスレッドが先頭ではありません: %+v", first.Items)
	}
	thread := first.Items[0]
	if thread.ReplyCount != 3 || thread.UnreadCount != 3 || len(thread.LatestReplies) != 2 {
		t.Errorf("件数が期待と異なります: reply=%d unread=%d latest=%d", thread.ReplyCount, thread.UnreadCount, len(thread.LatestReplies))
	}
	if thread.LatestReplies[0].Body != "b" || thread.LatestReplies[1].Body != "c" {
		t.Errorf("最新の返信が古い順に並んでいません: %s, %s", thread.LatestReplies[0].Body, thread.LatestReplies[1].Body)
	}
	if first.NextCursor == nil {
		t.Fatal("次ページのカーソルがありません")
	}

	input.Cursor = first.NextCursor
	second, err := repo.FindParticipatingThreads(ctx, input)
	if err != nil {
		t.Fatalf("取得に失敗しました: %v", err)
	}
	if len(second.Items) != 1 || second.Items[0].ThreadID != f.messages["mention"].ID.String() {
		t.Fatalf("2 ページ目が期待と異なります: %+v", second.Items)
	}
	if second.Items[0].UnreadCount != 0 {
		t.Errorf("自分の返信が未読に数えられています: %d", second.Items[0].UnreadCount)
	}
	if second.NextCursor != nil {
		t.Error("最後のページにカーソルがあります")
	}
}
