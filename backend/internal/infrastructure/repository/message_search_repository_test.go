package repository

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/migrate"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/database"
)

// TEST_DATABASE_URL に空の PostgreSQL を指定したときだけ実行する
func openTestClient(t *testing.T) *ent.Client {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL が未設定のためスキップします")
	}
	client, err := database.NewConnection(dsn)
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
		ch := client.Channel.Create().SetName(name).SetIsPrivate(private).SetWorkspaceID(f.workspaceID).SetCreatedBy(f.bob).SaveX(ctx)
		for _, m := range members {
			client.ChannelMember.Create().SetChannel(ch).SetUser(m).SaveX(ctx)
		}
		f.channels[name] = ch
	}
	newChannel("general", false, f.alice, f.bob)
	newChannel("dev", false, f.alice, f.bob)
	newChannel("dev/web", false, f.bob)
	newChannel("secret", true, f.bob)

	group := client.UserGroup.Create().SetName("designers").SetWorkspaceID(f.workspaceID).SetCreatedBy(f.bob).SaveX(ctx)
	client.UserGroupMember.Create().SetGroup(group).SetUser(f.alice).SaveX(ctx)

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
	newMessage("broadcast", 3, "dev", f.bob, "@channel 設計の締め切りと設計書")
	m5 := newMessage("group", 4, "general", f.bob, "group ping")
	client.MessageGroupMention.Create().SetMessage(m5).SetGroup(group).SaveX(ctx)
	reply := client.Message.Create().SetChannel(f.channels["general"]).SetUser(f.alice).SetParent(m1).SetBody("了解 設計").
		SetCreatedAt(base.Add(5 * time.Minute)).SaveX(ctx)
	f.messages["reply"] = reply
	m7 := newMessage("image", 6, "dev", f.bob, "スクショ")
	client.Attachment.Create().SetFileName("screenshot.png").SetMimeType("image/png").SetSizeBytes(1).SetStorageKey("k").
		SetMessage(m7).SetUploader(f.bob).SetChannel(f.channels["dev"]).SaveX(ctx)
	client.MessagePin.Create().SetChannel(f.channels["dev"]).SetMessage(m7).SetPinnedBy(f.bob).SaveX(ctx)
	m8 := newMessage("self", 7, "general", f.alice, "@alice self")
	client.MessageUserMention.Create().SetMessage(m8).SetUser(f.alice).SaveX(ctx)
	client.Message.Create().SetChannel(f.channels["general"]).SetUser(f.bob).SetBody("設計 deleted").
		SetCreatedAt(base.Add(8 * time.Minute)).SetDeletedAt(base.Add(9 * time.Minute)).SaveX(ctx)
	newMessage("not-broadcast", 9, "general", f.bob, "@channelx はメンションではない")
	newMessage("reverse", 10, "general", f.bob, "review of the design")
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

func TestSearchMessages(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	repo := NewMessageRepository(client)
	after := time.Date(2026, 9, 1, 0, 3, 0, 0, time.UTC)
	before := time.Date(2026, 9, 1, 0, 6, 0, 0, time.UTC)

	tests := []struct {
		name      string
		criteria  domainrepository.MessageSearchCriteria
		want      []string
		wantTotal int
	}{
		{name: "閲覧できない・削除済みを除いて新しい順", criteria: domainrepository.MessageSearchCriteria{Terms: []string{"設計"}}, want: []string{"reply", "broadcast", "mention"}},
		{name: "関連度順は出現回数の多いものが先", criteria: domainrepository.MessageSearchCriteria{Terms: []string{"設計"}, Sort: domainrepository.MessageSearchSortRelevance}, want: []string{"broadcast", "reply", "mention"}},
		{name: "句として含むものが先", criteria: domainrepository.MessageSearchCriteria{Terms: []string{"design", "review"}, Sort: domainrepository.MessageSearchSortRelevance}, want: []string{"link", "reverse"}},
		{name: "引用符やバックスラッシュを含む語でも関連度順にできる", criteria: domainrepository.MessageSearchCriteria{Terms: []string{"it's", `a\b'`}, Sort: domainrepository.MessageSearchSortRelevance}, want: []string{}},
		{name: "複数の語は AND", criteria: domainrepository.MessageSearchCriteria{Terms: []string{"設計", "締め切り"}}, want: []string{"broadcast"}},
		{name: "添付ファイル名にも一致する", criteria: domainrepository.MessageSearchCriteria{Terms: []string{"SCREENSHOT"}}, want: []string{"image"}},
		{name: "チャンネルとリンク", criteria: domainrepository.MessageSearchCriteria{ChannelIDs: []string{f.channels["dev"].ID.String(), f.channels["dev/web"].ID.String()}, Has: []domainrepository.MessageContentKind{domainrepository.MessageContentLink}}, want: []string{"link"}},
		{name: "画像", criteria: domainrepository.MessageSearchCriteria{Has: []domainrepository.MessageContentKind{domainrepository.MessageContentImage}}, want: []string{"image"}},
		{name: "画像以外のファイルはない", criteria: domainrepository.MessageSearchCriteria{Has: []domainrepository.MessageContentKind{domainrepository.MessageContentFile}}, want: []string{}},
		{name: "ピン留め", criteria: domainrepository.MessageSearchCriteria{PinnedOnly: true}, want: []string{"image"}},
		{name: "スレッド", criteria: domainrepository.MessageSearchCriteria{ThreadOnly: true}, want: []string{"reply", "mention"}},
		{name: "返信を除く", criteria: domainrepository.MessageSearchCriteria{Terms: []string{"設計"}, ExcludeReplies: true}, want: []string{"broadcast", "mention"}},
		{name: "自分宛て", criteria: domainrepository.MessageSearchCriteria{MentionsViewer: true}, want: []string{"self", "group", "broadcast", "mention"}},
		{name: "投稿者", criteria: domainrepository.MessageSearchCriteria{AuthorIDs: []string{f.alice.ID.String()}}, want: []string{"self", "reply"}},
		{name: "期間", criteria: domainrepository.MessageSearchCriteria{After: &after, Before: &before}, want: []string{"reply", "group", "broadcast"}},
		{name: "ページング", criteria: domainrepository.MessageSearchCriteria{Terms: []string{"設計"}, Limit: 1, Offset: 1}, want: []string{"broadcast"}, wantTotal: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.criteria
			c.WorkspaceID, c.ViewerID = f.workspaceID, f.alice.ID.String()
			if c.Limit == 0 {
				c.Limit = 20
			}
			messages, total, err := repo.SearchMessages(context.Background(), c)
			if err != nil {
				t.Fatalf("検索に失敗しました: %v", err)
			}
			if got := f.keys(t, messages); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("結果が期待と異なります: got=%v want=%v", got, tt.want)
			}
			wantTotal := tt.wantTotal
			if wantTotal == 0 {
				wantTotal = len(tt.want)
			}
			if total != wantTotal {
				t.Errorf("件数が期待と異なります: got=%d want=%d", total, wantTotal)
			}
		})
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

	input.CursorLastActivityAt = &first.NextCursor.LastActivityAt
	input.CursorThreadID = &first.NextCursor.ThreadID
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
