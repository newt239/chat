package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/message"
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

func (r *linkRepository) Create(ctx context.Context, link *entity.MessageLink) error {
	mid, err := utils.ParseUUID(link.MessageID, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	builder := client.MessageLink.Create().
		SetMessageID(mid).
		SetURL(link.URL)

	if link.ID != "" {
		linkID, err := utils.ParseUUID(link.ID, "link ID")
		if err != nil {
			return err
		}
		builder = builder.SetID(linkID)
	}

	if link.Title != nil {
		builder = builder.SetTitle(*link.Title)
	}

	if link.Description != nil {
		builder = builder.SetDescription(*link.Description)
	}

	if link.ImageURL != nil {
		builder = builder.SetImageURL(*link.ImageURL)
	}

	if link.SiteName != nil {
		builder = builder.SetSiteName(*link.SiteName)
	}

	if link.CardType != nil {
		builder = builder.SetCardType(*link.CardType)
	}

	ml, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	// Load edges
	ml, err = client.MessageLink.Query().
		Where(messagelink.ID(ml.ID)).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
		Only(ctx)
	if err != nil {
		return err
	}

	*link = *utils.MessageLinkToEntity(ml)
	return nil
}

func (r *linkRepository) FindByMessageID(ctx context.Context, messageID string) ([]*entity.MessageLink, error) {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	links, err := client.MessageLink.Query().
		Where(messagelink.HasMessageWith(message.ID(mid))).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
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

func (r *linkRepository) FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageLink, error) {
	if len(messageIDs) == 0 {
		return []*entity.MessageLink{}, nil
	}

	// Parse all message IDs
	parsedIDs := make([]uuid.UUID, 0, len(messageIDs))
	for _, id := range messageIDs {
		parsedID, err := utils.ParseUUID(id, "message ID")
		if err != nil {
			return nil, err
		}
		parsedIDs = append(parsedIDs, parsedID)
	}

	client := transaction.ResolveClient(ctx, r.client)
	links, err := client.MessageLink.Query().
		Where(messagelink.HasMessageWith(message.IDIn(parsedIDs...))).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
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
	client := transaction.ResolveClient(ctx, r.client)
	ml, err := client.MessageLink.Query().
		Where(messagelink.URL(url)).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return utils.MessageLinkToEntity(ml), nil
}

func (r *linkRepository) DeleteByMessageID(ctx context.Context, messageID string) error {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	_, err = client.MessageLink.Delete().
		Where(messagelink.HasMessageWith(message.ID(mid))).
		Exec(ctx)

	return err
}
