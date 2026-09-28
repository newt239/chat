package link

import (
	"context"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/ogp"
)

const (
	channelID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	messageID = "f1111111-1111-1111-1111-111111111111"
)

type stubOGPService struct {
	fetched []string
}

func (s *stubOGPService) FetchOGP(_ context.Context, url string) (*entity.OGPData, error) {
	s.fetched = append(s.fetched, url)
	title := "取得したタイトル"
	return &entity.OGPData{Title: &title}, nil
}

func (s *stubOGPService) ExtractURLs(text string) []string {
	return ogp.ExtractURLs(text)
}

type stubLinkRepo struct {
	repository.MessageLinkRepository
	cached map[string]*entity.MessageLink
}

func (r *stubLinkRepo) FindByURL(_ context.Context, url string) (*entity.MessageLink, error) {
	return r.cached[url], nil
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

func TestProcessLinks(t *testing.T) {
	cachedTitle := "保存済みのタイトル"
	ogpService := &stubOGPService{}
	service := NewLinkProcessingService(
		ogpService,
		&stubLinkRepo{cached: map[string]*entity.MessageLink{
			"https://cached.example.com/": {OGP: entity.OGPData{Title: &cachedTitle}},
		}},
		&stubMessageRepo{},
		&stubChannelRepo{},
	)
	permalink := "http://localhost:5173/app/general/" + channelID + "?message=" + messageID
	otherWorkspace := "http://localhost:5173/app/other/" + channelID + "?message=" + messageID
	missing := "http://localhost:5173/app/general/" + channelID + "?message=f9999999-9999-9999-9999-999999999999"
	body := permalink + " " + otherWorkspace + " " + missing + " https://cached.example.com/ https://new.example.com/"

	links, err := service.ProcessLinks(context.Background(), body, "general")
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	if links[0].LinkedMessageID == nil || *links[0].LinkedMessageID != messageID || links[0].OGP.Title != nil {
		t.Errorf("同じワークスペースのメッセージリンクとして扱われていません: %+v", links[0])
	}
	if links[1].LinkedMessageID != nil {
		t.Error("別のワークスペースの URL をメッセージリンクとして扱っています")
	}
	if links[2].LinkedMessageID != nil || links[2].OGP.Title != nil {
		t.Error("存在しないメッセージの URL にリンク先や OGP が付いています")
	}
	if links[3].OGP.Title == nil || *links[3].OGP.Title != cachedTitle {
		t.Error("保存済みの OGP が再利用されていません")
	}
	if len(ogpService.fetched) != 2 || ogpService.fetched[0] != otherWorkspace || ogpService.fetched[1] != "https://new.example.com/" {
		t.Errorf("OGP を取得した URL が期待と異なります: %v", ogpService.fetched)
	}
}
