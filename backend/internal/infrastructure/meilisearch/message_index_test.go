package meilisearch

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	meili "github.com/meilisearch/meilisearch-go"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

func TestBuildFilter(t *testing.T) {
	after := time.UnixMilli(1000)
	got := buildFilter(domainrepository.MessageSearchCriteria{
		WorkspaceID:    `w"s`,
		ChannelIDs:     []string{"c1", "c2"},
		AuthorIDs:      []string{"u2"},
		Has:            []domainrepository.MessageContentKind{domainrepository.MessageContentImage, domainrepository.MessageContentLink},
		PinnedOnly:     true,
		ThreadOnly:     true,
		ExcludeReplies: true,
		Mention:        &domainrepository.MessageSearchScope{UserID: "u1", GroupIDs: []string{"g1"}, JoinedChannelIDs: []string{"c1"}},
		After:          &after,
	})
	want := `workspace_id = "w\"s" AND channel_id IN ["c1", "c2"] AND sender_id IN ["u2"] AND has = "image" AND has = "link"` +
		` AND pinned = true AND (parent_id IS NOT NULL OR has_replies = true) AND parent_id IS NULL` +
		` AND (mentioned_user_ids = "u1" OR mentioned_group_ids IN ["g1"] OR (mentions_channel = true AND channel_id IN ["c1"]))` +
		` AND created_at >= 1000`
	if got != want {
		t.Errorf("フィルタが期待と異なります:\n got=%s\nwant=%s", got, want)
	}
}

// TEST_MEILISEARCH_URL（と TEST_MEILISEARCH_API_KEY）に Meilisearch を指定したときだけ実行する
func TestMessageIndexSearch(t *testing.T) {
	url := os.Getenv("TEST_MEILISEARCH_URL")
	if url == "" {
		t.Skip("TEST_MEILISEARCH_URL が未設定のためスキップします")
	}
	ctx := context.Background()
	apiKey := os.Getenv("TEST_MEILISEARCH_API_KEY")
	uid := fmt.Sprintf("test_messages_%d", time.Now().UnixNano())
	client := meili.New(url, meili.WithAPIKey(apiKey))
	index := &MessageIndex{index: client.Index(uid)}
	t.Cleanup(func() { _, _ = client.DeleteIndexWithContext(ctx, uid) })
	if err := index.EnsureSettings(ctx); err != nil {
		t.Fatalf("設定に失敗しました: %v", err)
	}

	base := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	parent := "m1"
	docs := []domainrepository.MessageSearchDocument{
		{ID: "m1", ChannelID: "general", SenderID: "bob", Body: "設計レビューをお願いします", MentionedUserIDs: []string{"alice"}, HasReplies: true},
		{ID: "m2", ChannelID: "dev", SenderID: "bob", Body: "Design review 資料 https://example.com", Has: []domainrepository.MessageContentKind{domainrepository.MessageContentLink}},
		{ID: "m3", ChannelID: "secret", SenderID: "bob", Body: "設計 secret"},
		{ID: "m4", ChannelID: "dev", SenderID: "bob", Body: "@channel 設計の締め切りと設計書", MentionsChannel: true},
		{ID: "m5", ChannelID: "general", SenderID: "alice", ParentID: &parent, Body: "了解 設計"},
		{ID: "m6", ChannelID: "dev", SenderID: "bob", Body: "スクショ", AttachmentNames: []string{"screenshot.png"}, Pinned: true,
			Has: []domainrepository.MessageContentKind{domainrepository.MessageContentImage}},
		{ID: "m7", ChannelID: "general", SenderID: "bob", Body: "group ping", MentionedGroupIDs: []string{"designers"}},
	}
	for i := range docs {
		docs[i].WorkspaceID = "ws"
		docs[i].CreatedAt = base.Add(time.Duration(i) * time.Minute)
	}
	if err := index.Upsert(ctx, docs); err != nil {
		t.Fatalf("登録に失敗しました: %v", err)
	}
	if err := index.Delete(ctx, []string{"m7"}); err != nil {
		t.Fatalf("削除に失敗しました: %v", err)
	}
	waitForIndexing(t, index)

	viewable := []string{"general", "dev"}
	after := base.Add(3 * time.Minute)
	tests := []struct {
		name     string
		criteria domainrepository.MessageSearchCriteria
		want     []string
	}{
		{name: "閲覧できないチャンネルを除いて新しい順", criteria: domainrepository.MessageSearchCriteria{Terms: []string{"設計"}}, want: []string{"m5", "m4", "m1"}},
		{name: "複数の語は AND", criteria: domainrepository.MessageSearchCriteria{Terms: []string{"設計", "締め切り"}}, want: []string{"m4"}},
		{name: "英語は大文字小文字を区別しない", criteria: domainrepository.MessageSearchCriteria{Terms: []string{"DESIGN"}}, want: []string{"m2"}},
		{name: "添付ファイル名にも一致する", criteria: domainrepository.MessageSearchCriteria{Terms: []string{"screenshot"}}, want: []string{"m6"}},
		{name: "リンク", criteria: domainrepository.MessageSearchCriteria{Has: []domainrepository.MessageContentKind{domainrepository.MessageContentLink}}, want: []string{"m2"}},
		{name: "ピン留め", criteria: domainrepository.MessageSearchCriteria{PinnedOnly: true}, want: []string{"m6"}},
		{name: "スレッド", criteria: domainrepository.MessageSearchCriteria{ThreadOnly: true}, want: []string{"m5", "m1"}},
		{name: "返信を除く", criteria: domainrepository.MessageSearchCriteria{Terms: []string{"設計"}, ExcludeReplies: true}, want: []string{"m4", "m1"}},
		{name: "投稿者", criteria: domainrepository.MessageSearchCriteria{AuthorIDs: []string{"alice"}}, want: []string{"m5"}},
		{name: "期間", criteria: domainrepository.MessageSearchCriteria{After: &after}, want: []string{"m6", "m5", "m4"}},
		{
			name: "自分宛て",
			criteria: domainrepository.MessageSearchCriteria{Mention: &domainrepository.MessageSearchScope{
				UserID: "alice", JoinedChannelIDs: []string{"dev"}, GroupIDs: []string{"designers"},
			}},
			want: []string{"m4", "m1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.criteria
			c.WorkspaceID, c.ChannelIDs, c.Page, c.PerPage = "ws", viewable, 1, 20
			res, err := index.Search(ctx, c)
			if err != nil {
				t.Fatalf("検索に失敗しました: %v", err)
			}
			if !reflect.DeepEqual(res.MessageIDs, tt.want) || res.Total != len(tt.want) {
				t.Errorf("結果が期待と異なります: got=%v total=%d want=%v", res.MessageIDs, res.Total, tt.want)
			}
		})
	}

	t.Run("ページング", func(t *testing.T) {
		res, err := index.Search(ctx, domainrepository.MessageSearchCriteria{WorkspaceID: "ws", ChannelIDs: viewable, Terms: []string{"設計"}, Page: 2, PerPage: 1})
		if err != nil {
			t.Fatalf("検索に失敗しました: %v", err)
		}
		if !reflect.DeepEqual(res.MessageIDs, []string{"m4"}) || res.Total != 3 {
			t.Errorf("2 ページ目が期待と異なります: %v total=%d", res.MessageIDs, res.Total)
		}
	})
}

// waitForIndexing は登録・削除のタスクがすべて処理されるまで待ちます
func waitForIndexing(t *testing.T, index *MessageIndex) {
	t.Helper()
	for range 100 {
		stats, err := index.index.GetStatsWithContext(context.Background(), nil)
		if err != nil {
			t.Fatalf("状態の取得に失敗しました: %v", err)
		}
		if !stats.IsIndexing && stats.NumberOfDocuments == 6 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("インデックスの更新が終わりません")
}
