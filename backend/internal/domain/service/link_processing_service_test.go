package service

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/domain/repository"
)

const (
	channelID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	messageID = "f1111111-1111-1111-1111-111111111111"
)

type stubOGPService struct {
	mu      sync.Mutex
	fetched []string
}

func (s *stubOGPService) FetchOGP(_ context.Context, url string) (*entity.OGPData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fetched = append(s.fetched, url)
	title := "取得したタイトル"
	return &entity.OGPData{Title: &title}, nil
}

type stubLinkRepo struct {
	repository.MessageLinkRepository
	previews map[string]*entity.LinkPreview
}

func (r *stubLinkRepo) FindPreviewsByURLs(_ context.Context, urls []string) (map[string]*entity.LinkPreview, error) {
	result := map[string]*entity.LinkPreview{}
	for _, url := range urls {
		if p, ok := r.previews[url]; ok {
			result[url] = p
		}
	}
	return result, nil
}

func (r *stubLinkRepo) UpsertPreview(_ context.Context, p *entity.LinkPreview) error {
	p.ID = "preview-" + p.URL
	r.previews[p.URL] = p
	return nil
}

type stubMessageRepo struct {
	repository.MessageRepository
}

func (r *stubMessageRepo) FindByID(_ context.Context, id string) (*entity.Message, error) {
	if id != messageID {
		return nil, nil
	}
	return &entity.Message{ID: messageID, ChannelID: channelID}, nil
}

type stubChannelRepo struct {
	repository.ChannelRepository
}

func (r *stubChannelRepo) FindByID(_ context.Context, id string) (*entity.Channel, error) {
	return &entity.Channel{ID: id, WorkspaceID: "general"}, nil
}

func TestPrepareLinks(t *testing.T) {
	cachedTitle := "保存済みのタイトル"
	ogpService := &stubOGPService{}
	linkRepo := &stubLinkRepo{previews: map[string]*entity.LinkPreview{
		"https://cached.example.com/": {ID: "cached", URL: "https://cached.example.com/", OGP: entity.OGPData{Title: &cachedTitle}, FetchedAt: time.Now()},
		"https://stale.example.com/":  {ID: "stale", URL: "https://stale.example.com/", OGP: entity.OGPData{Title: &cachedTitle}, FetchedAt: time.Now().Add(-2 * entity.LinkPreviewTTL)},
	}}
	service := NewLinkProcessingService(ogpService, linkRepo, &stubMessageRepo{}, &stubChannelRepo{})
	permalink := "http://localhost:5173/app/general/" + channelID + "?message=" + messageID
	otherWorkspace := "http://localhost:5173/app/other/" + channelID + "?message=" + messageID
	missing := "http://localhost:5173/app/general/" + channelID + "?message=f9999999-9999-9999-9999-999999999999"
	body := permalink + " " + otherWorkspace + " " + missing + " https://cached.example.com/ https://new.example.com/ https://stale.example.com/"

	links, err := service.PrepareLinks(context.Background(), body, "general")
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	if links[0].LinkedMessageID == nil || *links[0].LinkedMessageID != messageID || links[0].LinkPreviewID != nil {
		t.Errorf("同じワークスペースのメッセージリンクとして扱われていません: %+v", links[0])
	}
	if links[1].LinkedMessageID != nil || links[1].LinkPreviewID == nil {
		t.Error("別のワークスペースの URL をメッセージリンクとして扱っています")
	}
	if links[2].LinkedMessageID != nil || links[2].LinkPreviewID != nil {
		t.Error("存在しないメッセージの URL にリンク先やプレビューが付いています")
	}
	if *links[3].LinkPreviewID != "cached" || *links[3].OGP.Title != cachedTitle {
		t.Error("新しいプレビューが再利用されていません")
	}
	slices.Sort(ogpService.fetched)
	if want := []string{otherWorkspace, "https://new.example.com/", "https://stale.example.com/"}; !slices.Equal(ogpService.fetched, want) {
		t.Errorf("OGP を取得した URL が期待と異なります: %v", ogpService.fetched)
	}
}
