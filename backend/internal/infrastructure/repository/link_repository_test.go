package repository

import (
	"context"
	"testing"
	"time"

	"github.com/newt239/chat/ent/linkpreview"
	"github.com/newt239/chat/internal/domain/entity"
)

func TestLinkRepositorySharesPreviewByURL(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	repo := NewLinkRepository(client)
	ctx := context.Background()
	title, url := "動画", "https://www.youtube.com/watch?v="+f.workspaceID

	for _, videoID := range []string{"old", "abc"} {
		preview := &entity.LinkPreview{URL: url, OGP: entity.OGPData{Title: &title, YouTube: &entity.YouTubeVideo{VideoID: videoID}}, FetchedAt: time.Now()}
		if err := repo.UpsertPreview(ctx, preview); err != nil {
			t.Fatalf("プレビューを保存できません: %v", err)
		}
		links := []*entity.MessageLink{{MessageID: f.messages["mention"].ID.String(), URL: url, LinkPreviewID: &preview.ID}}
		if videoID == "abc" {
			links[0].MessageID = f.messages["link"].ID.String()
		}
		if err := repo.CreateBulk(ctx, links); err != nil || links[0].ID == "" {
			t.Fatalf("リンクを保存できません: %v", err)
		}
	}

	if n := client.LinkPreview.Query().Where(linkpreview.URL(url)).CountX(ctx); n != 1 {
		t.Errorf("同じ URL のプレビューは 1 行にまとめることを期待しましたが %d 行でした", n)
	}
	links, err := repo.FindByMessageIDs(ctx, []string{f.messages["mention"].ID.String(), f.messages["link"].ID.String()})
	if err != nil || len(links) != 2 {
		t.Fatalf("取得に失敗しました: %v %v", links, err)
	}
	for _, l := range links {
		if l.OGP.Title == nil || *l.OGP.Title != title || l.OGP.YouTube == nil || l.OGP.YouTube.VideoID != "abc" {
			t.Errorf("上書きした OGP が復元されていません: %+v", l.OGP)
		}
	}
	previews, err := repo.FindPreviewsByURLs(ctx, []string{url, "https://missing.example.com/"})
	if err != nil || len(previews) != 1 || previews[url].OGP.YouTube == nil {
		t.Errorf("URL からプレビューを取得できません: %+v %v", previews, err)
	}
}
