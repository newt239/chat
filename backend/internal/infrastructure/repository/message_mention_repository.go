package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/messagegroupmention"
	"github.com/newt239/chat/ent/messageusermention"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type messageMentionRepository struct {
	client *ent.Client
}

func NewMessageMentionRepository(client *ent.Client) domainrepository.MessageMentionRepository {
	return &messageMentionRepository{client: client}
}

func (r *messageMentionRepository) FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageUserMention, []*entity.MessageGroupMention, error) {
	ids, err := parseUUIDs(messageIDs, "message ID")
	if err != nil {
		return nil, nil, err
	}
	client := transaction.ResolveClient(ctx, r.client)
	users, err := client.MessageUserMention.Query().Where(messageusermention.MessageIDIn(ids...)).All(ctx)
	if err != nil {
		return nil, nil, err
	}
	groups, err := client.MessageGroupMention.Query().Where(messagegroupmention.MessageIDIn(ids...)).All(ctx)
	if err != nil {
		return nil, nil, err
	}
	userMentions := convertAll(users, func(m *ent.MessageUserMention) *entity.MessageUserMention {
		return &entity.MessageUserMention{MessageID: m.MessageID.String(), UserID: m.UserID.String(), ViaGroupID: optionalString(m.ViaGroupID)}
	})
	groupMentions := convertAll(groups, func(m *ent.MessageGroupMention) *entity.MessageGroupMention {
		return &entity.MessageGroupMention{MessageID: m.MessageID.String(), GroupID: m.GroupID.String()}
	})
	return userMentions, groupMentions, nil
}

func (r *messageMentionRepository) Create(ctx context.Context, users []*entity.MessageUserMention, groups []*entity.MessageGroupMention) error {
	client := transaction.ResolveClient(ctx, r.client)
	userBuilders := make([]*ent.MessageUserMentionCreate, 0, len(users))
	for _, m := range users {
		mid, err := parseUUID(m.MessageID, "message ID")
		if err != nil {
			return err
		}
		uid, err := parseUUID(m.UserID, "user ID")
		if err != nil {
			return err
		}
		userBuilders = append(userBuilders, client.MessageUserMention.Create().SetMessageID(mid).SetUserID(uid).SetNillableViaGroupID(parseUUIDPtr(m.ViaGroupID)))
	}
	groupBuilders := make([]*ent.MessageGroupMentionCreate, 0, len(groups))
	for _, m := range groups {
		mid, err := parseUUID(m.MessageID, "message ID")
		if err != nil {
			return err
		}
		gid, err := parseUUID(m.GroupID, "group ID")
		if err != nil {
			return err
		}
		groupBuilders = append(groupBuilders, client.MessageGroupMention.Create().SetMessageID(mid).SetGroupID(gid))
	}
	if len(userBuilders) > 0 {
		if err := client.MessageUserMention.CreateBulk(userBuilders...).Exec(ctx); err != nil {
			return err
		}
	}
	if len(groupBuilders) > 0 {
		return client.MessageGroupMention.CreateBulk(groupBuilders...).Exec(ctx)
	}
	return nil
}

func (r *messageMentionRepository) DeleteByMessageID(ctx context.Context, messageID string) error {
	mid, err := parseUUID(messageID, "message ID")
	if err != nil {
		return err
	}
	client := transaction.ResolveClient(ctx, r.client)
	if _, err := client.MessageUserMention.Delete().Where(messageusermention.MessageID(mid)).Exec(ctx); err != nil {
		return err
	}
	_, err = client.MessageGroupMention.Delete().Where(messagegroupmention.MessageID(mid)).Exec(ctx)
	return err
}
