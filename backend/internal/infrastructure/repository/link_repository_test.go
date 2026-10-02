package repository

import (
	"context"
	"testing"

	"github.com/newt239/chat/ent/linkpreview"
	"github.com/newt239/chat/internal/domain/entity"
)

func TestLinkRepositorySharesPreviewByURL(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	repo := NewLinkRepository(client)
	ctx := context.Background()
	title, url := "動画", "https://www.youtube.com/watch?v="+f.workspaceID

	for _, key := range []string{"mention", "link"} {
		link := &entity.MessageLink{MessageID: f.messages[key].ID.String(), URL: url, OGP: entity.OGPData{Title: &title, YouTube: &entity.YouTubeVideo{VideoID: "abc"}}}
		if err := repo.Create(ctx, link); err != nil {
			t.Fatalf("保存に失敗しました: %v", err)
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
			t.Errorf("OGP が復元されていません: %+v", l.OGP)
		}
	}
	cached, err := repo.FindByURL(ctx, url)
	if err != nil || cached == nil || cached.OGP.YouTube == nil {
		t.Errorf("URL からプレビューを取得できません: %+v %v", cached, err)
	}
}
