package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/systemmessage"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type systemMessageRepository struct {
	client *ent.Client
}

func NewSystemMessageRepository(client *ent.Client) domainrepository.SystemMessageRepository {
	return &systemMessageRepository{client: client}
}

func (r *systemMessageRepository) Create(ctx context.Context, msg *entity.SystemMessage) error {
	channelID, err := parseUUID(msg.ChannelID, "channel ID")
	if err != nil {
		return err
	}
	sm, err := transaction.ResolveClient(ctx, r.client).SystemMessage.Create().
		SetChannelID(channelID).
		SetKind(string(msg.Kind)).
		SetPayload(msg.Payload).
		SetNillableActorID(parseUUIDPtr(msg.ActorID)).
		Save(ctx)
	if err != nil {
		return err
	}
	msg.ID = sm.ID.String()
	msg.CreatedAt = sm.CreatedAt
	return nil
}

func (r *systemMessageRepository) FindByChannelIDs(ctx context.Context, channelIDs []string, limit int, since *time.Time, until *time.Time, ascending bool) ([]*entity.SystemMessage, error) {
	chIDs, err := parseUUIDs(channelIDs, "channel ID")
	if err != nil {
		return nil, err
	}
	query := transaction.ResolveClient(ctx, r.client).SystemMessage.Query().Where(systemmessage.ChannelIDIn(chIDs...))
	if since != nil {
		query.Where(systemmessage.CreatedAtGT(*since))
	}
	if until != nil {
		query.Where(systemmessage.CreatedAtLT(*until))
	}
	if limit > 0 {
		query.Limit(limit)
	}
	order := ent.Desc(systemmessage.FieldCreatedAt)
	if ascending {
		order = ent.Asc(systemmessage.FieldCreatedAt)
	}
	rows, err := query.Order(order).All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(rows, func(sm *ent.SystemMessage) *entity.SystemMessage {
		return &entity.SystemMessage{
			ID:        sm.ID.String(),
			ChannelID: sm.ChannelID.String(),
			Kind:      entity.SystemMessageKind(sm.Kind),
			Payload:   sm.Payload,
			ActorID:   optionalString(sm.ActorID),
			CreatedAt: sm.CreatedAt,
		}
	}), nil
}
