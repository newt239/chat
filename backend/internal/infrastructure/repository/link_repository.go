package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/linkpreview"
	"github.com/newt239/chat/ent/linkpreviewxpost"
	"github.com/newt239/chat/ent/linkpreviewyoutube"
	"github.com/newt239/chat/ent/messagelink"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type linkRepository struct {
	client *ent.Client
}

func NewLinkRepository(client *ent.Client) domainrepository.MessageLinkRepository {
	return &linkRepository{client: client}
}

// CreateBulk はリンクをまとめて保存します。プレビューは UpsertPreview で先に保存しておく
func (r *linkRepository) CreateBulk(ctx context.Context, links []*entity.MessageLink) error {
	if len(links) == 0 {
		return nil
	}
	client := transaction.ResolveClient(ctx, r.client)
	builders := make([]*ent.MessageLinkCreate, 0, len(links))
	for _, link := range links {
		mid, err := utils.ParseUUID(link.MessageID, "message ID")
		if err != nil {
			return err
		}
		builders = append(builders, client.MessageLink.Create().
			SetMessageID(mid).
			SetURL(link.URL).
			SetNillableLinkPreviewID(utils.ParseUUIDPtr(link.LinkPreviewID)).
			SetNillableLinkedMessageID(utils.ParseUUIDPtr(link.LinkedMessageID)).
			SetNillableCreatedAt(nonZeroTime(link.CreatedAt)))
	}
	saved, err := client.MessageLink.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return err
	}
	for i, ml := range saved {
		links[i].ID = ml.ID.String()
		links[i].CreatedAt = ml.CreatedAt
	}
	return nil
}

func (r *linkRepository) FindPreviewsByURLs(ctx context.Context, urls []string) (map[string]*entity.LinkPreview, error) {
	result := make(map[string]*entity.LinkPreview, len(urls))
	if len(urls) == 0 {
		return result, nil
	}
	previews, err := transaction.ResolveClient(ctx, r.client).LinkPreview.Query().
		Where(linkpreview.URLIn(urls...)).
		WithYoutube().
		WithXPost().
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, lp := range previews {
		result[lp.URL] = &entity.LinkPreview{ID: lp.ID.String(), URL: lp.URL, OGP: utils.LinkPreviewToOGP(lp), FetchedAt: lp.FetchedAt}
	}
	return result, nil
}

func (r *linkRepository) UpsertPreview(ctx context.Context, preview *entity.LinkPreview) error {
	return transaction.WithTx(ctx, r.client, func(client *ent.Client) error {
		ogp := preview.OGP
		previewID, err := client.LinkPreview.Create().
			SetURL(preview.URL).
			SetNillableTitle(ogp.Title).
			SetNillableDescription(ogp.Description).
			SetNillableImageURL(ogp.ImageURL).
			SetNillableSiteName(ogp.SiteName).
			SetNillableCardType(ogp.CardType).
			SetNillableImageWidth(ogp.ImageWidth).
			SetNillableImageHeight(ogp.ImageHeight).
			SetFetchedAt(preview.FetchedAt).
			OnConflictColumns(linkpreview.FieldURL).
			UpdateNewValues().
			ID(ctx)
		if err != nil {
			return err
		}

		if _, err := client.LinkPreviewYoutube.Delete().Where(linkpreviewyoutube.LinkPreviewID(previewID)).Exec(ctx); err != nil {
			return err
		}
		if yt := ogp.YouTube; yt != nil {
			if err := client.LinkPreviewYoutube.Create().
				SetLinkPreviewID(previewID).
				SetVideoID(yt.VideoID).
				SetNillableChannelName(yt.ChannelName).
				SetNillableDurationSeconds(yt.DurationSeconds).
				Exec(ctx); err != nil {
				return err
			}
		}
		if _, err := client.LinkPreviewXPost.Delete().Where(linkpreviewxpost.LinkPreviewID(previewID)).Exec(ctx); err != nil {
			return err
		}
		if x := ogp.XPost; x != nil {
			if err := client.LinkPreviewXPost.Create().
				SetLinkPreviewID(previewID).
				SetAuthorName(x.AuthorName).
				SetAuthorHandle(x.AuthorHandle).
				Exec(ctx); err != nil {
				return err
			}
		}
		preview.ID = previewID.String()
		return nil
	})
}

func (r *linkRepository) FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageLink, error) {
	if len(messageIDs) == 0 {
		return []*entity.MessageLink{}, nil
	}
	parsedIDs, err := utils.ParseUUIDs(messageIDs, "message ID")
	if err != nil {
		return nil, err
	}

	links, err := transaction.ResolveClient(ctx, r.client).MessageLink.Query().
		Where(messagelink.MessageIDIn(parsedIDs...)).
		WithLinkPreview(func(q *ent.LinkPreviewQuery) { q.WithYoutube().WithXPost() }).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.MessageLink, 0, len(links))
	for _, ml := range links {
		result = append(result, utils.MessageLinkToEntity(ml))
	}
	return result, nil
}

func (r *linkRepository) DeleteByMessageID(ctx context.Context, messageID string) error {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	_, err = transaction.ResolveClient(ctx, r.client).MessageLink.Delete().
		Where(messagelink.MessageID(mid)).
		Exec(ctx)
	return err
}

// nonZeroTime は未設定の日時を nil にして DB の既定値を使わせます
func nonZeroTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
