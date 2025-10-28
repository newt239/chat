package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/message"
	"github.com/newt239/chat/ent/messageusermention"
	"github.com/newt239/chat/ent/user"
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

func (r *messageUserMentionRepository) FindByMessageID(ctx context.Context, messageID string) ([]*entity.MessageUserMention, error) {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	mentions, err := client.MessageUserMention.Query().
		Where(messageusermention.HasMessageWith(message.ID(mid))).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
		WithUser().
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
		Where(messageusermention.HasMessageWith(message.IDIn(parsedIDs...))).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
		WithUser().
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

func (r *messageUserMentionRepository) FindByUserID(ctx context.Context, userID string, limit int, since *time.Time) ([]*entity.MessageUserMention, error) {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	query := client.MessageUserMention.Query().
		Where(messageusermention.HasUserWith(user.ID(uid)))

	if since != nil {
		query = query.Where(messageusermention.CreatedAtGT(*since))
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	mentions, err := query.
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
		WithUser().
		Order(ent.Desc(messageusermention.FieldCreatedAt)).
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

func (r *messageUserMentionRepository) Create(ctx context.Context, mention *entity.MessageUserMention) error {
	mid, err := utils.ParseUUID(mention.MessageID, "message ID")
	if err != nil {
		return err
	}

	uid, err := utils.ParseUUID(mention.UserID, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	_, err = client.MessageUserMention.Create().
		SetMessageID(mid).
		SetUserID(uid).
		Save(ctx)

	return err
}

func (r *messageUserMentionRepository) DeleteByMessageID(ctx context.Context, messageID string) error {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	_, err = client.MessageUserMention.Delete().
		Where(messageusermention.HasMessageWith(message.ID(mid))).
		Exec(ctx)

	return err
}
