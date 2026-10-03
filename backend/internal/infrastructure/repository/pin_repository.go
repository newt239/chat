package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/messagepin"
	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
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

// Create は同じメッセージが既にピン留めされていれば ErrPinExists を返します
func (r *pinRepository) Create(ctx context.Context, pin *entity.MessagePin) error {
	mp, err := transaction.ResolveClient(ctx, r.client).MessagePin.Create().
		SetChannelID(utils.ParseUUIDOrNil(pin.ChannelID)).
		SetMessageID(utils.ParseUUIDOrNil(pin.MessageID)).
		SetPinnedByID(utils.ParseUUIDOrNil(pin.PinnedBy)).
		Save(ctx)
	if ent.IsConstraintError(err) {
		return domerr.ErrPinExists
	}
	if err != nil {
		return err
	}
	pin.ID = mp.ID.String()
	pin.PinnedAt = mp.CreatedAt
	return nil
}

func (r *pinRepository) Delete(ctx context.Context, channelID, messageID string) error {
	chID := utils.ParseUUIDOrNil(channelID)
	msgID := utils.ParseUUIDOrNil(messageID)

	client := transaction.ResolveClient(ctx, r.client)
	_, err := client.MessagePin.Delete().
		Where(
			messagepin.ChannelID(chID),
			messagepin.MessageID(msgID),
		).
		Exec(ctx)
	return err
}

func (r *pinRepository) List(ctx context.Context, channelID string, limit int, cursor *string) ([]*entity.MessagePin, *string, error) {
	chID := utils.ParseUUIDOrNil(channelID)
	client := transaction.ResolveClient(ctx, r.client)

	q := client.MessagePin.Query().
		Where(messagepin.ChannelID(chID)).
		WithMessage().
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

func (r *pinRepository) FindByMessageIDs(ctx context.Context, messageIDs []string) (map[string]*entity.MessagePin, error) {
	ids := make([]uuid.UUID, 0, len(messageIDs))
	for _, id := range messageIDs {
		ids = append(ids, utils.ParseUUIDOrNil(id))
	}

	rows, err := transaction.ResolveClient(ctx, r.client).MessagePin.Query().
		Where(messagepin.MessageIDIn(ids...)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	pins := make(map[string]*entity.MessagePin, len(rows))
	for _, mp := range rows {
		pin := messagePinToEntity(mp)
		pins[pin.MessageID] = pin
	}
	return pins, nil
}

func messagePinToEntity(mp *ent.MessagePin) *entity.MessagePin {
	return &entity.MessagePin{
		ID:        mp.ID.String(),
		ChannelID: mp.ChannelID.String(),
		MessageID: mp.MessageID.String(),
		PinnedBy:  mp.PinnedByID.String(),
		PinnedAt:  mp.CreatedAt,
		Message:   utils.MessageToEntity(mp.Edges.Message),
	}
}
