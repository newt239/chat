package link

import (
	"context"
	"sync"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

// maxConcurrentFetches は 1 つのメッセージで同時に OGP を取りに行く上限
const maxConcurrentFetches = 4

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

// PrepareLinks は同じワークスペースのメッセージ URL を引用にし、それ以外は保存済みのプレビューを使います
// プレビューがないか古い URL だけを並行して取り直し、取得に失敗しても空のプレビューを保存して取り直しを抑えます
func (s *linkProcessingService) PrepareLinks(ctx context.Context, body, workspaceID string) ([]*entity.MessageLink, error) {
	urls := s.ogpService.ExtractURLs(body)
	links := make([]*entity.MessageLink, 0, len(urls))
	var external []*entity.MessageLink
	for _, url := range urls {
		link := &entity.MessageLink{URL: url}
		links = append(links, link)
		if permalink, ok := entity.ParseMessagePermalink(url); ok && permalink.WorkspaceID == workspaceID {
			linkedMessageID, err := s.resolvePermalink(ctx, permalink)
			if err != nil {
				return nil, err
			}
			link.LinkedMessageID = linkedMessageID
			continue
		}
		external = append(external, link)
	}
	if len(external) == 0 {
		return links, nil
	}

	externalURLs := make([]string, len(external))
	for i, link := range external {
		externalURLs[i] = link.URL
	}
	previews, err := s.linkRepo.FindPreviewsByURLs(ctx, externalURLs)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var stale []string
	for _, url := range externalURLs {
		if p := previews[url]; p == nil || p.IsStale(now) {
			stale = append(stale, url)
		}
	}
	for _, preview := range s.fetchPreviews(ctx, stale, now) {
		if err := s.linkRepo.UpsertPreview(ctx, preview); err != nil {
			return nil, err
		}
		previews[preview.URL] = preview
	}

	for _, link := range external {
		preview := previews[link.URL]
		link.LinkPreviewID = &preview.ID
		link.OGP = preview.OGP
	}
	return links, nil
}

// fetchPreviews は URL の OGP を並行して取得します。失敗した URL は空の OGP にします
func (s *linkProcessingService) fetchPreviews(ctx context.Context, urls []string, now time.Time) []*entity.LinkPreview {
	previews := make([]*entity.LinkPreview, len(urls))
	sem := make(chan struct{}, maxConcurrentFetches)
	var wg sync.WaitGroup
	for i, url := range urls {
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()
			preview := &entity.LinkPreview{URL: url, FetchedAt: now}
			if ogp, err := s.ogpService.FetchOGP(ctx, url); err == nil {
				preview.OGP = *ogp
			}
			previews[i] = preview
		})
	}
	wg.Wait()
	return previews
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
