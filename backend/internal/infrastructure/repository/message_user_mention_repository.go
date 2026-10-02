package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/messageusermention"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type messageUserMentionRepository struct {
	client *ent.Client
}

func NewMessageUserMentionRepository(client *ent.Client) domainrepository.MessageUserMentionRepository {
	return &messageUserMentionRepository{client: client}
}

func (r *messageUserMentionRepository) FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageUserMention, error) {
	if len(messageIDs) == 0 {
		return []*entity.MessageUserMention{}, nil
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
	mentions, err := client.MessageUserMention.Query().
		Where(messageusermention.MessageIDIn(parsedIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.MessageUserMention, 0, len(mentions))
	for _, mum := range mentions {
		result = append(result, utils.MessageUserMentionToEntity(mum))
	}

	return result, nil
}

func (r *messageUserMentionRepository) CreateBulk(ctx context.Context, mentions []*entity.MessageUserMention) error {
	if len(mentions) == 0 {
		return nil
	}
	client := transaction.ResolveClient(ctx, r.client)
	builders := make([]*ent.MessageUserMentionCreate, 0, len(mentions))
	for _, mention := range mentions {
		mid, err := utils.ParseUUID(mention.MessageID, "message ID")
		if err != nil {
			return err
		}
		uid, err := utils.ParseUUID(mention.UserID, "user ID")
		if err != nil {
			return err
		}
		builders = append(builders, client.MessageUserMention.Create().
			SetMessageID(mid).
			SetUserID(uid).
			SetNillableViaGroupID(utils.ParseUUIDPtr(mention.ViaGroupID)))
	}
	return client.MessageUserMention.CreateBulk(builders...).Exec(ctx)
}

func (r *messageUserMentionRepository) DeleteByMessageID(ctx context.Context, messageID string) error {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	_, err = client.MessageUserMention.Delete().
		Where(messageusermention.MessageID(mid)).
		Exec(ctx)

	return err
}
