package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channellink"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type channelLinkRepository struct {
	client *ent.Client
}

func NewChannelLinkRepository(client *ent.Client) domainrepository.ChannelLinkRepository {
	return &channelLinkRepository{client: client}
}

func (r *channelLinkRepository) FindByID(ctx context.Context, id string) (*entity.ChannelLink, error) {
	linkID, err := parseUUID(id, "link ID")
	if err != nil {
		return nil, err
	}

	link, err := orNil(transaction.ResolveClient(ctx, r.client).ChannelLink.Get(ctx, linkID))
	if link == nil {
		return nil, err
	}
	return channelLinkToEntity(link), nil
}

func (r *channelLinkRepository) FindByChannelID(ctx context.Context, channelID string) ([]*entity.ChannelLink, error) {
	cid, err := parseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}

	links, err := transaction.ResolveClient(ctx, r.client).ChannelLink.Query().
		Where(channellink.ChannelID(cid)).
		Order(ent.Asc(channellink.FieldPosition), ent.Asc(channellink.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(links, channelLinkToEntity), nil
}

func (r *channelLinkRepository) Create(ctx context.Context, link *entity.ChannelLink) error {
	cid, err := parseUUID(link.ChannelID, "channel ID")
	if err != nil {
		return err
	}
	uid, err := parseUUID(link.CreatedBy, "user ID")
	if err != nil {
		return err
	}

	created, err := transaction.ResolveClient(ctx, r.client).ChannelLink.Create().
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
	return nil
}

func (r *channelLinkRepository) Update(ctx context.Context, link *entity.ChannelLink) error {
	linkID, err := parseUUID(link.ID, "link ID")
	if err != nil {
		return err
	}

	return transaction.ResolveClient(ctx, r.client).ChannelLink.UpdateOneID(linkID).
		SetTitle(link.Title).
		SetURL(link.URL).
		Exec(ctx)
}

func (r *channelLinkRepository) Delete(ctx context.Context, id string) error {
	linkID, err := parseUUID(id, "link ID")
	if err != nil {
		return err
	}

	return transaction.ResolveClient(ctx, r.client).ChannelLink.DeleteOneID(linkID).Exec(ctx)
}

func (r *channelLinkRepository) UpdatePositions(ctx context.Context, linkIDs []string) error {
	client := transaction.ResolveClient(ctx, r.client)
	for position, id := range linkIDs {
		linkID, err := parseUUID(id, "link ID")
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
	return &entity.ChannelLink{
		ID:        link.ID.String(),
		Title:     link.Title,
		URL:       link.URL,
		Position:  link.Position,
		ChannelID: link.ChannelID.String(),
		CreatedBy: link.CreatedByID.String(),
	}
}
