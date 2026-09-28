package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/channellink"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type channelLinkRepository struct {
	client *ent.Client
}

func NewChannelLinkRepository(client *ent.Client) domainrepository.ChannelLinkRepository {
	return &channelLinkRepository{client: client}
}

func (r *channelLinkRepository) FindByID(ctx context.Context, id string) (*entity.ChannelLink, error) {
	linkID, err := utils.ParseUUID(id, "link ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	link, err := client.ChannelLink.Query().
		Where(channellink.ID(linkID)).
		WithChannel().
		WithCreatedBy().
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return channelLinkToEntity(link), nil
}

func (r *channelLinkRepository) FindByChannelID(ctx context.Context, channelID string) ([]*entity.ChannelLink, error) {
	cid, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	links, err := client.ChannelLink.Query().
		Where(channellink.HasChannelWith(channel.ID(cid))).
		WithChannel().
		WithCreatedBy().
		Order(ent.Asc(channellink.FieldPosition), ent.Asc(channellink.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.ChannelLink, 0, len(links))
	for _, link := range links {
		result = append(result, channelLinkToEntity(link))
	}
	return result, nil
}

func (r *channelLinkRepository) Create(ctx context.Context, link *entity.ChannelLink) error {
	cid, err := utils.ParseUUID(link.ChannelID, "channel ID")
	if err != nil {
		return err
	}
	uid, err := utils.ParseUUID(link.CreatedBy, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	created, err := client.ChannelLink.Create().
		SetChannelID(cid).
		SetCreatedByID(uid).
		SetTitle(link.Title).
		SetURL(link.URL).
		SetPosition(link.Position).
		Save(ctx)
	if err != nil {
		return err
	}

	link.ID = created.ID.String()
	link.CreatedAt = created.CreatedAt
	link.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *channelLinkRepository) Update(ctx context.Context, link *entity.ChannelLink) error {
	linkID, err := utils.ParseUUID(link.ID, "link ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	updated, err := client.ChannelLink.UpdateOneID(linkID).
		SetTitle(link.Title).
		SetURL(link.URL).
		Save(ctx)
	if err != nil {
		return err
	}

	link.UpdatedAt = updated.UpdatedAt
	return nil
}

func (r *channelLinkRepository) Delete(ctx context.Context, id string) error {
	linkID, err := utils.ParseUUID(id, "link ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	return client.ChannelLink.DeleteOneID(linkID).Exec(ctx)
}

func (r *channelLinkRepository) UpdatePositions(ctx context.Context, linkIDs []string) error {
	client := transaction.ResolveClient(ctx, r.client)
	for position, id := range linkIDs {
		linkID, err := utils.ParseUUID(id, "link ID")
		if err != nil {
			return err
		}
		if err := client.ChannelLink.UpdateOneID(linkID).SetPosition(position).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func channelLinkToEntity(link *ent.ChannelLink) *entity.ChannelLink {
	result := &entity.ChannelLink{
		ID:        link.ID.String(),
		Title:     link.Title,
		URL:       link.URL,
		Position:  link.Position,
		CreatedAt: link.CreatedAt,
		UpdatedAt: link.UpdatedAt,
	}
	if link.Edges.Channel != nil {
		result.ChannelID = link.Edges.Channel.ID.String()
	}
	if link.Edges.CreatedBy != nil {
		result.CreatedBy = link.Edges.CreatedBy.ID.String()
	}
	return result
}
