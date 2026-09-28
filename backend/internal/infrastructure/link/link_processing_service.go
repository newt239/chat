package link

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

type linkProcessingService struct {
	ogpService  service.OGPService
	linkRepo    repository.MessageLinkRepository
	messageRepo repository.MessageRepository
	channelRepo repository.ChannelRepository
}

func NewLinkProcessingService(
	ogpService service.OGPService,
	linkRepo repository.MessageLinkRepository,
	messageRepo repository.MessageRepository,
	channelRepo repository.ChannelRepository,
) service.LinkProcessingService {
	return &linkProcessingService{
		ogpService:  ogpService,
		linkRepo:    linkRepo,
		messageRepo: messageRepo,
		channelRepo: channelRepo,
	}
}

// ProcessLinks は本文中の URL を抽出します。同じワークスペースのメッセージ URL は OGP を取得せず、
// それ以外は保存済みの OGP を再利用し、なければ取得します。取得に失敗してもリンクは残します
func (s *linkProcessingService) ProcessLinks(ctx context.Context, body, workspaceID string) ([]*entity.MessageLink, error) {
	urls := s.ogpService.ExtractURLs(body)
	links := make([]*entity.MessageLink, 0, len(urls))

	for _, urlStr := range urls {
		link := &entity.MessageLink{URL: urlStr}
		links = append(links, link)

		if permalink, ok := entity.ParseMessagePermalink(urlStr); ok && permalink.WorkspaceID == workspaceID {
			linkedMessageID, err := s.resolvePermalink(ctx, permalink)
			if err != nil {
				return nil, err
			}
			link.LinkedMessageID = linkedMessageID
			continue
		}

		if cached, err := s.linkRepo.FindByURL(ctx, urlStr); err == nil && cached != nil && cached.OGP.Title != nil {
			link.OGP = cached.OGP
			continue
		}
		if ogpData, err := s.ogpService.FetchOGP(ctx, urlStr); err == nil {
			link.OGP = *ogpData
		}
	}

	return links, nil
}

// resolvePermalink はメッセージリンクが同じワークスペースに実在するメッセージを指す場合にその ID を返します
func (s *linkProcessingService) resolvePermalink(ctx context.Context, permalink entity.MessagePermalink) (*string, error) {
	msg, err := s.messageRepo.FindByID(ctx, permalink.MessageID)
	if err != nil || msg == nil || msg.ChannelID != permalink.ChannelID {
		return nil, err
	}
	ch, err := s.channelRepo.FindByID(ctx, msg.ChannelID)
	if err != nil || ch == nil || ch.WorkspaceID != permalink.WorkspaceID {
		return nil, err
	}
	return &msg.ID, nil
}
