package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/message"
	"github.com/newt239/chat/ent/messagepin"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type pinRepository struct {
	client *ent.Client
}

func NewPinRepository(client *ent.Client) domainrepository.PinRepository {
	return &pinRepository{client: client}
}

func (r *pinRepository) Create(ctx context.Context, pin *entity.MessagePin) error {
	chID := utils.ParseUUIDOrNil(pin.ChannelID)
	msgID := utils.ParseUUIDOrNil(pin.MessageID)
	byID := utils.ParseUUIDOrNil(pin.PinnedBy)

	client := transaction.ResolveClient(ctx, r.client)

	mp, err := client.MessagePin.Create().
		SetChannelID(chID).
		SetMessageID(msgID).
		SetPinnedByID(byID).
		Save(ctx)
	if err != nil {
		return err
	}

	// 再読込してエッジを付与
	mp, err = client.MessagePin.Query().
		Where(messagepin.IDEQ(mp.ID)).
		WithChannel().
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) { q2.WithWorkspace().WithCreatedBy() }).WithUser()
		}).
		WithPinnedBy().
		Only(ctx)
	if err != nil {
		return err
	}

	*pin = *messagePinToEntity(mp)
	return nil
}

func (r *pinRepository) Delete(ctx context.Context, channelID, messageID string) error {
	chID := utils.ParseUUIDOrNil(channelID)
	msgID := utils.ParseUUIDOrNil(messageID)

	client := transaction.ResolveClient(ctx, r.client)
	_, err := client.MessagePin.Delete().
		Where(
			messagepin.HasChannelWith(channel.ID(chID)),
			messagepin.HasMessageWith(message.ID(msgID)),
		).
		Exec(ctx)
	return err
}

func (r *pinRepository) List(ctx context.Context, channelID string, limit int, cursor *string) ([]*entity.MessagePin, *string, error) {
	chID := utils.ParseUUIDOrNil(channelID)
	client := transaction.ResolveClient(ctx, r.client)

	q := client.MessagePin.Query().
		Where(messagepin.HasChannelWith(channel.ID(chID))).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) { q2.WithWorkspace().WithCreatedBy() }).WithUser()
		}).
		WithPinnedBy().
		Order(ent.Desc(messagepin.FieldCreatedAt)).
		Limit(limit + 1)

	if cursor != nil && *cursor != "" {
		if t, err := time.Parse(time.RFC3339, *cursor); err == nil {
			q = q.Where(messagepin.CreatedAtLT(t))
		}
	}

	rows, err := q.All(ctx)
	if err != nil {
		return nil, nil, err
	}

	var next *string
	if len(rows) > limit {
		t := rows[limit].CreatedAt.Format(time.RFC3339)
		next = &t
		rows = rows[:limit]
	}

	out := make([]*entity.MessagePin, 0, len(rows))
	for _, mp := range rows {
		out = append(out, messagePinToEntity(mp))
	}
	return out, next, nil
}

func messagePinToEntity(mp *ent.MessagePin) *entity.MessagePin {
	var channelID, messageID, pinnedBy string
	if mp.Edges.Channel != nil {
		channelID = mp.Edges.Channel.ID.String()
	}
	if mp.Edges.Message != nil {
		messageID = mp.Edges.Message.ID.String()
	}
	if mp.Edges.PinnedBy != nil {
		pinnedBy = mp.Edges.PinnedBy.ID.String()
	}
	return &entity.MessagePin{
		ID:        mp.ID.String(),
		ChannelID: channelID,
		MessageID: messageID,
		PinnedBy:  pinnedBy,
		PinnedAt:  mp.CreatedAt,
		Message:   utils.MessageToEntity(mp.Edges.Message),
	}
}
