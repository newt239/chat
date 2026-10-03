package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/message"
	"github.com/newt239/chat/ent/messagepin"
	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type pinRepository struct {
	client *ent.Client
}

func NewPinRepository(client *ent.Client) domainrepository.PinRepository {
	return &pinRepository{client: client}
}

func (r *pinRepository) Create(ctx context.Context, pin *entity.MessagePin) error {
	channelID, err := parseUUID(pin.ChannelID, "channel ID")
	if err != nil {
		return err
	}
	messageID, err := parseUUID(pin.MessageID, "message ID")
	if err != nil {
		return err
	}
	pinnedBy, err := parseUUID(pin.PinnedBy, "user ID")
	if err != nil {
		return err
	}
	mp, err := transaction.ResolveClient(ctx, r.client).MessagePin.Create().
		SetChannelID(channelID).
		SetMessageID(messageID).
		SetPinnedByID(pinnedBy).
		Save(ctx)
	if ent.IsConstraintError(err) {
		return domerr.ErrPinExists
	}
	if err != nil {
		return err
	}
	pin.PinnedAt = mp.CreatedAt
	return nil
}

func (r *pinRepository) Delete(ctx context.Context, channelID, messageID string) error {
	chID, err := parseUUID(channelID, "channel ID")
	if err != nil {
		return err
	}
	msgID, err := parseUUID(messageID, "message ID")
	if err != nil {
		return err
	}
	_, err = transaction.ResolveClient(ctx, r.client).MessagePin.Delete().
		Where(messagepin.ChannelID(chID), messagepin.MessageID(msgID)).
		Exec(ctx)
	return err
}

func (r *pinRepository) List(ctx context.Context, channelID string, limit int) ([]*entity.MessagePin, error) {
	chID, err := parseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}
	rows, err := transaction.ResolveClient(ctx, r.client).MessagePin.Query().
		Where(messagepin.ChannelID(chID), messagepin.HasMessageWith(message.DeletedAtIsNil())).
		WithMessage().
		Order(ent.Desc(messagepin.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(rows, messagePinToEntity), nil
}

func (r *pinRepository) FindByMessageIDs(ctx context.Context, messageIDs []string) (map[string]*entity.MessagePin, error) {
	ids, err := parseUUIDs(messageIDs, "message ID")
	if err != nil {
		return nil, err
	}
	rows, err := transaction.ResolveClient(ctx, r.client).MessagePin.Query().
		Where(messagepin.MessageIDIn(ids...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	pins := make(map[string]*entity.MessagePin, len(rows))
	for _, mp := range rows {
		pins[mp.MessageID.String()] = messagePinToEntity(mp)
	}
	return pins, nil
}

func messagePinToEntity(mp *ent.MessagePin) *entity.MessagePin {
	return &entity.MessagePin{
		ChannelID: mp.ChannelID.String(),
		MessageID: mp.MessageID.String(),
		PinnedBy:  mp.PinnedByID.String(),
		PinnedAt:  mp.CreatedAt,
		Message:   messageToEntity(mp.Edges.Message),
	}
}
