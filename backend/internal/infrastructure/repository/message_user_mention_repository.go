package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/messageusermention"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type messageUserMentionRepository struct {
	client *ent.Client
}

func NewMessageUserMentionRepository(client *ent.Client) domainrepository.MessageUserMentionRepository {
	return &messageUserMentionRepository{client: client}
}

func (r *messageUserMentionRepository) FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageUserMention, error) {
	parsedIDs, err := parseUUIDs(messageIDs, "message ID")
	if err != nil {
		return nil, err
	}
	client := transaction.ResolveClient(ctx, r.client)
	mentions, err := client.MessageUserMention.Query().
		Where(messageusermention.MessageIDIn(parsedIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return convertAll(mentions, func(m *ent.MessageUserMention) *entity.MessageUserMention {
		return &entity.MessageUserMention{MessageID: m.MessageID.String(), UserID: m.UserID.String(), ViaGroupID: optionalString(m.ViaGroupID)}
	}), nil
}

func (r *messageUserMentionRepository) CreateBulk(ctx context.Context, mentions []*entity.MessageUserMention) error {
	if len(mentions) == 0 {
		return nil
	}
	client := transaction.ResolveClient(ctx, r.client)
	builders := make([]*ent.MessageUserMentionCreate, 0, len(mentions))
	for _, mention := range mentions {
		mid, err := parseUUID(mention.MessageID, "message ID")
		if err != nil {
			return err
		}
		uid, err := parseUUID(mention.UserID, "user ID")
		if err != nil {
			return err
		}
		builders = append(builders, client.MessageUserMention.Create().
			SetMessageID(mid).
			SetUserID(uid).
			SetNillableViaGroupID(parseUUIDPtr(mention.ViaGroupID)))
	}
	return client.MessageUserMention.CreateBulk(builders...).Exec(ctx)
}

func (r *messageUserMentionRepository) DeleteByMessageID(ctx context.Context, messageID string) error {
	mid, err := parseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	_, err = client.MessageUserMention.Delete().
		Where(messageusermention.MessageID(mid)).
		Exec(ctx)

	return err
}
