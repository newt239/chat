package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

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

// Create はリンクを保存します。メッセージへのリンクでなければ OGP を URL ごとの link_preview に上書きで保存し、そこから参照します
func (r *linkRepository) Create(ctx context.Context, link *entity.MessageLink) error {
	mid, err := utils.ParseUUID(link.MessageID, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	builder := client.MessageLink.Create().
		SetMessageID(mid).
		SetURL(link.URL).
		SetNillableLinkedMessageID(utils.ParseUUIDPtr(link.LinkedMessageID))
	if link.ID != "" {
		linkID, err := utils.ParseUUID(link.ID, "link ID")
		if err != nil {
			return err
		}
		builder = builder.SetID(linkID)
	}
	if link.LinkedMessageID == nil {
		previewID, err := r.savePreview(ctx, client, link.URL, link.OGP)
		if err != nil {
			return err
		}
		builder = builder.SetLinkPreviewID(previewID)
	}

	ml, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	link.ID = ml.ID.String()
	link.CreatedAt = ml.CreatedAt
	return nil
}

func (r *linkRepository) savePreview(ctx context.Context, client *ent.Client, url string, ogp entity.OGPData) (uuid.UUID, error) {
	previewID, err := client.LinkPreview.Create().
		SetURL(url).
		SetNillableTitle(ogp.Title).
		SetNillableDescription(ogp.Description).
		SetNillableImageURL(ogp.ImageURL).
		SetNillableSiteName(ogp.SiteName).
		SetNillableCardType(ogp.CardType).
		SetNillableImageWidth(ogp.ImageWidth).
		SetNillableImageHeight(ogp.ImageHeight).
		SetFetchedAt(time.Now()).
		OnConflictColumns(linkpreview.FieldURL).
		UpdateNewValues().
		ID(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	if _, err := client.LinkPreviewYoutube.Delete().Where(linkpreviewyoutube.LinkPreviewID(previewID)).Exec(ctx); err != nil {
		return uuid.Nil, err
	}
	if yt := ogp.YouTube; yt != nil {
		if err := client.LinkPreviewYoutube.Create().
			SetLinkPreviewID(previewID).
			SetVideoID(yt.VideoID).
			SetNillableChannelName(yt.ChannelName).
			SetNillableDurationSeconds(yt.DurationSeconds).
			Exec(ctx); err != nil {
			return uuid.Nil, err
		}
	}
	if _, err := client.LinkPreviewXPost.Delete().Where(linkpreviewxpost.LinkPreviewID(previewID)).Exec(ctx); err != nil {
		return uuid.Nil, err
	}
	if x := ogp.XPost; x != nil {
		if err := client.LinkPreviewXPost.Create().
			SetLinkPreviewID(previewID).
			SetAuthorName(x.AuthorName).
			SetAuthorHandle(x.AuthorHandle).
			Exec(ctx); err != nil {
			return uuid.Nil, err
		}
	}
	return previewID, nil
}

func (r *linkRepository) FindByMessageID(ctx context.Context, messageID string) ([]*entity.MessageLink, error) {
	return r.FindByMessageIDs(ctx, []string{messageID})
}

func (r *linkRepository) FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageLink, error) {
	if len(messageIDs) == 0 {
		return []*entity.MessageLink{}, nil
	}
	parsedIDs, err := parseUUIDs(messageIDs, "message ID")
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

func (r *linkRepository) FindByURL(ctx context.Context, url string) (*entity.MessageLink, error) {
	lp, err := transaction.ResolveClient(ctx, r.client).LinkPreview.Query().
		Where(linkpreview.URL(url)).
		WithYoutube().
		WithXPost().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &entity.MessageLink{URL: lp.URL, OGP: utils.LinkPreviewToOGP(lp)}, nil
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
