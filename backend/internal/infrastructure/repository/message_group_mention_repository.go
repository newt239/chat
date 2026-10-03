package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/messagegroupmention"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type messageGroupMentionRepository struct {
	client *ent.Client
}

func NewMessageGroupMentionRepository(client *ent.Client) domainrepository.MessageGroupMentionRepository {
	return &messageGroupMentionRepository{client: client}
}

func (r *messageGroupMentionRepository) FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageGroupMention, error) {
	parsedIDs, err := parseUUIDs(messageIDs, "message ID")
	if err != nil {
		return nil, err
	}
	client := transaction.ResolveClient(ctx, r.client)
	mentions, err := client.MessageGroupMention.Query().
		Where(messagegroupmention.MessageIDIn(parsedIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return convertAll(mentions, func(m *ent.MessageGroupMention) *entity.MessageGroupMention {
		return &entity.MessageGroupMention{MessageID: m.MessageID.String(), GroupID: m.GroupID.String()}
	}), nil
}

func (r *messageGroupMentionRepository) CreateBulk(ctx context.Context, mentions []*entity.MessageGroupMention) error {
	if len(mentions) == 0 {
		return nil
	}
	client := transaction.ResolveClient(ctx, r.client)
	builders := make([]*ent.MessageGroupMentionCreate, 0, len(mentions))
	for _, mention := range mentions {
		mid, err := parseUUID(mention.MessageID, "message ID")
		if err != nil {
			return err
		}
		gid, err := parseUUID(mention.GroupID, "group ID")
		if err != nil {
			return err
		}
		builders = append(builders, client.MessageGroupMention.Create().SetMessageID(mid).SetGroupID(gid))
	}
	return client.MessageGroupMention.CreateBulk(builders...).Exec(ctx)
}

func (r *messageGroupMentionRepository) DeleteByMessageID(ctx context.Context, messageID string) error {
	mid, err := parseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	_, err = client.MessageGroupMention.Delete().
		Where(messagegroupmention.MessageID(mid)).
		Exec(ctx)

	return err
}
