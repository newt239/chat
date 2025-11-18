package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/systemmessage"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type systemMessageRepository struct {
	client *ent.Client
}

func NewSystemMessageRepository(client *ent.Client) domainrepository.SystemMessageRepository {
	return &systemMessageRepository{client: client}
}

func (r *systemMessageRepository) Create(ctx context.Context, msg *entity.SystemMessage) error {
	chID := utils.ParseUUIDOrNil(msg.ChannelID)
	actorIDPtr := utils.ParseUUIDPtrOrNil(msg.ActorID)

	client := transaction.ResolveClient(ctx, r.client)

	builder := client.SystemMessage.Create().
		SetChannelID(chID).
		SetKind(string(msg.Kind)).
		SetPayload(msg.Payload)

	if actorIDPtr != nil {
		builder = builder.SetActorID(*actorIDPtr)
	}

	sm, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	// 再読込（必要に応じてエッジを付与）
	_, _ = client.SystemMessage.Query().
		Where(systemmessage.IDEQ(sm.ID)).
		WithChannel(func(q *ent.ChannelQuery) { q.WithWorkspace().WithCreatedBy() }).
		WithActor().
		Only(ctx)

	msg.ID = sm.ID.String()
	msg.CreatedAt = sm.CreatedAt
	return nil
}

func (r *systemMessageRepository) FindByChannelID(ctx context.Context, channelID string, limit int, since *time.Time, until *time.Time) ([]*entity.SystemMessage, error) {
	chID, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	q := client.SystemMessage.Query().
		Where(systemmessage.HasChannelWith(channel.ID(chID)))

	if since != nil {
		q = q.Where(systemmessage.CreatedAtGT(*since))
	}
	if until != nil {
		q = q.Where(systemmessage.CreatedAtLT(*until))
	}
	if limit > 0 {
		q = q.Limit(limit)
	}

	rows, err := q.
		WithChannel().
		WithActor().
		Order(ent.Desc(systemmessage.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]*entity.SystemMessage, 0, len(rows))
	for _, sm := range rows {
		var actorID *string
		if sm.Edges.Actor != nil {
			s := sm.Edges.Actor.ID.String()
			actorID = &s
		}
		chID := ""
		if sm.Edges.Channel != nil {
			chID = sm.Edges.Channel.ID.String()
		}
		out = append(out, &entity.SystemMessage{
			ID:        sm.ID.String(),
			ChannelID: chID,
			Kind:      entity.SystemMessageKind(sm.Kind),
			Payload:   sm.Payload,
			ActorID:   actorID,
			CreatedAt: sm.CreatedAt,
		})
	}
	return out, nil
}
