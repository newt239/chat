package link

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

type linkProcessingService struct {
	ogpService service.OGPService
	linkRepo   repository.MessageLinkRepository
}

func NewLinkProcessingService(
	ogpService service.OGPService,
	linkRepo repository.MessageLinkRepository,
) service.LinkProcessingService {
	return &linkProcessingService{
		ogpService: ogpService,
		linkRepo:   linkRepo,
	}
}

// ProcessLinks はメッセージ本文からリンクを抽出し、OGP情報を取得します
func (s *linkProcessingService) ProcessLinks(ctx context.Context, body string) ([]*entity.MessageLink, error) {
	// URLを抽出
	urls := s.ogpService.ExtractURLs(body)

	var links []*entity.MessageLink

	for _, urlStr := range urls {
		// OGP情報を取得
		ogpData, err := s.ogpService.FetchOGP(ctx, urlStr)
		if err != nil {
			// OGP取得に失敗してもリンクは作成
			link := &entity.MessageLink{
				URL:         urlStr,
				Title:       nil,
				Description: nil,
				ImageURL:    nil,
				SiteName:    nil,
				CardType:    nil,
			}
			links = append(links, link)
			continue
		}

		// OGP情報を含むリンクを作成
		link := &entity.MessageLink{
			URL:         urlStr,
			Title:       ogpData.Title,
			Description: ogpData.Description,
			ImageURL:    ogpData.ImageURL,
			SiteName:    ogpData.SiteName,
			CardType:    ogpData.CardType,
		}
		links = append(links, link)
	}

	return links, nil
}
