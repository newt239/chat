package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/ent"
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

	msg.ID = sm.ID.String()
	msg.CreatedAt = sm.CreatedAt
	return nil
}

func (r *systemMessageRepository) FindByChannelIDs(ctx context.Context, channelIDs []string, limit int, since *time.Time, until *time.Time, ascending bool) ([]*entity.SystemMessage, error) {
	chIDs, err := utils.ParseUUIDs(channelIDs, "channel ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	q := client.SystemMessage.Query().
		Where(
			systemmessage.ChannelIDIn(chIDs...),
			// 退出・削除は以前は記録していたがタイムラインには出さない
			systemmessage.KindNotIn(string(entity.SystemMessageKindMemberLeft), string(entity.SystemMessageKindMemberRemoved)),
		)

	if since != nil {
		q = q.Where(systemmessage.CreatedAtGT(*since))
	}
	if until != nil {
		q = q.Where(systemmessage.CreatedAtLT(*until))
	}
	if limit > 0 {
		q = q.Limit(limit)
	}

	order := ent.Desc(systemmessage.FieldCreatedAt)
	if ascending {
		order = ent.Asc(systemmessage.FieldCreatedAt)
	}
	rows, err := q.
		Order(order).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]*entity.SystemMessage, 0, len(rows))
	for _, sm := range rows {
		out = append(out, &entity.SystemMessage{
			ID:        sm.ID.String(),
			ChannelID: sm.ChannelID.String(),
			Kind:      entity.SystemMessageKind(sm.Kind),
			Payload:   sm.Payload,
			ActorID:   utils.UUIDPtrToStringPtr(sm.ActorID),
			CreatedAt: sm.CreatedAt,
		})
	}
	return out, nil
}
