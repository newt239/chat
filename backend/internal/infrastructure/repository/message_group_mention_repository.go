package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/messagegroupmention"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type messageGroupMentionRepository struct {
	client *ent.Client
}

func NewMessageGroupMentionRepository(client *ent.Client) domainrepository.MessageGroupMentionRepository {
	return &messageGroupMentionRepository{client: client}
}

func (r *messageGroupMentionRepository) FindByMessageID(ctx context.Context, messageID string) ([]*entity.MessageGroupMention, error) {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	mentions, err := client.MessageGroupMention.Query().
		Where(messagegroupmention.MessageID(mid)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.MessageGroupMention, 0, len(mentions))
	for _, mgm := range mentions {
		result = append(result, utils.MessageGroupMentionToEntity(mgm))
	}

	return result, nil
}

func (r *messageGroupMentionRepository) FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageGroupMention, error) {
	if len(messageIDs) == 0 {
		return []*entity.MessageGroupMention{}, nil
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
	mentions, err := client.MessageGroupMention.Query().
		Where(messagegroupmention.MessageIDIn(parsedIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.MessageGroupMention, 0, len(mentions))
	for _, mgm := range mentions {
		result = append(result, utils.MessageGroupMentionToEntity(mgm))
	}

	return result, nil
}

func (r *messageGroupMentionRepository) FindByGroupID(ctx context.Context, groupID string, limit int, since *time.Time) ([]*entity.MessageGroupMention, error) {
	gid, err := utils.ParseUUID(groupID, "group ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	query := client.MessageGroupMention.Query().
		Where(messagegroupmention.GroupID(gid))

	if since != nil {
		query = query.Where(messagegroupmention.CreatedAtGT(*since))
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	mentions, err := query.
		Order(ent.Desc(messagegroupmention.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.MessageGroupMention, 0, len(mentions))
	for _, mgm := range mentions {
		result = append(result, utils.MessageGroupMentionToEntity(mgm))
	}

	return result, nil
}

func (r *messageGroupMentionRepository) Create(ctx context.Context, mention *entity.MessageGroupMention) error {
	mid, err := utils.ParseUUID(mention.MessageID, "message ID")
	if err != nil {
		return err
	}

	gid, err := utils.ParseUUID(mention.GroupID, "group ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	_, err = client.MessageGroupMention.Create().
		SetMessageID(mid).
		SetGroupID(gid).
		Save(ctx)

	return err
}

func (r *messageGroupMentionRepository) DeleteByMessageID(ctx context.Context, messageID string) error {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	_, err = client.MessageGroupMention.Delete().
		Where(messagegroupmention.MessageID(mid)).
		Exec(ctx)

	return err
}
