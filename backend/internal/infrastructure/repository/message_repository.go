package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/message"
	"github.com/newt239/chat/ent/messagereaction"
	"github.com/newt239/chat/ent/user"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type messageRepository struct {
	client *ent.Client
}

func NewMessageRepository(client *ent.Client) domainrepository.MessageRepository {
	return &messageRepository{client: client}
}

func (r *messageRepository) FindByID(ctx context.Context, id string) (*entity.Message, error) {
	messageID, err := utils.ParseUUID(id, "message ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	m, err := client.Message.Query().
		Where(message.ID(messageID)).
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		WithUser().
		WithParent().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return utils.MessageToEntity(m), nil
}

func (r *messageRepository) FindByChannelID(ctx context.Context, channelID string, limit int, since *time.Time, until *time.Time) ([]*entity.Message, error) {
	chID, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	query := client.Message.Query().
		Where(
			message.HasChannelWith(channel.ID(chID)),
			message.Not(message.HasParent()),
			message.DeletedAtIsNil(),
		)

	if since != nil {
		query = query.Where(message.CreatedAtGT(*since))
	}

	if until != nil {
		query = query.Where(message.CreatedAtLT(*until))
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	messages, err := query.
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		WithUser().
		WithParent().
		Order(ent.Desc(message.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.Message, 0, len(messages))
	for _, m := range messages {
		result = append(result, utils.MessageToEntity(m))
	}

	return result, nil
}

func (r *messageRepository) FindByChannelIDIncludingDeleted(ctx context.Context, channelID string, limit int, since *time.Time, until *time.Time) ([]*entity.Message, error) {
	chID, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	query := client.Message.Query().
		Where(
			message.HasChannelWith(channel.ID(chID)),
			message.Not(message.HasParent()),
		)

	if since != nil {
		query = query.Where(message.CreatedAtGT(*since))
	}

	if until != nil {
		query = query.Where(message.CreatedAtLT(*until))
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	messages, err := query.
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		WithUser().
		WithParent().
		Order(ent.Desc(message.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.Message, 0, len(messages))
	for _, m := range messages {
		result = append(result, utils.MessageToEntity(m))
	}

	return result, nil
}

func (r *messageRepository) FindThreadReplies(ctx context.Context, parentID string) ([]*entity.Message, error) {
	pID, err := utils.ParseUUID(parentID, "parent ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	messages, err := client.Message.Query().
		Where(
			message.HasParentWith(message.ID(pID)),
			message.DeletedAtIsNil(),
		).
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		WithUser().
		WithParent().
		Order(ent.Asc(message.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.Message, 0, len(messages))
	for _, m := range messages {
		result = append(result, utils.MessageToEntity(m))
	}

	return result, nil
}

func (r *messageRepository) FindThreadRepliesIncludingDeleted(ctx context.Context, parentID string) ([]*entity.Message, error) {
	pID, err := utils.ParseUUID(parentID, "parent ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	messages, err := client.Message.Query().
		Where(message.HasParentWith(message.ID(pID))).
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		WithUser().
		WithParent().
		Order(ent.Asc(message.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.Message, 0, len(messages))
	for _, m := range messages {
		result = append(result, utils.MessageToEntity(m))
	}

	return result, nil
}

func (r *messageRepository) Create(ctx context.Context, msg *entity.Message) error {
	channelID, err := utils.ParseUUID(msg.ChannelID, "channel ID")
	if err != nil {
		return err
	}

	userID, err := utils.ParseUUID(msg.UserID, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	builder := client.Message.Create().
		SetChannelID(channelID).
		SetUserID(userID).
		SetBody(msg.Body)

	if msg.ID != "" {
		messageID, err := utils.ParseUUID(msg.ID, "message ID")
		if err != nil {
			return err
		}
		builder = builder.SetID(messageID)
	}

	if msg.ParentID != nil {
		parentID, err := utils.ParseUUID(*msg.ParentID, "parent ID")
		if err != nil {
			return err
		}
		builder = builder.SetParentID(parentID)
	}

	if msg.EditedAt != nil {
		builder = builder.SetEditedAt(*msg.EditedAt)
	}

	if msg.DeletedAt != nil {
		builder = builder.SetDeletedAt(*msg.DeletedAt)
	}

	if msg.DeletedBy != nil {
		deletedBy, err := utils.ParseUUID(*msg.DeletedBy, "deleted_by user ID")
		if err != nil {
			return err
		}
		builder = builder.SetDeletedBy(deletedBy)
	}

	m, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	// Load edges
	m, err = client.Message.Query().
		Where(message.ID(m.ID)).
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		WithUser().
		WithParent().
		Only(ctx)
	if err != nil {
		return err
	}

	*msg = *utils.MessageToEntity(m)
	return nil
}

func (r *messageRepository) Update(ctx context.Context, msg *entity.Message) error {
	messageID, err := utils.ParseUUID(msg.ID, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	builder := client.Message.UpdateOneID(messageID).
		SetBody(msg.Body)

	if msg.EditedAt != nil {
		builder = builder.SetEditedAt(*msg.EditedAt)
	}

	m, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	// Load edges
	m, err = client.Message.Query().
		Where(message.ID(m.ID)).
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		WithUser().
		WithParent().
		Only(ctx)
	if err != nil {
		return err
	}

	if !m.EditedAt.IsZero() {
		msg.EditedAt = &m.EditedAt
	}
	return nil
}

func (r *messageRepository) Delete(ctx context.Context, id string) error {
	messageID, err := utils.ParseUUID(id, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	return client.Message.DeleteOneID(messageID).Exec(ctx)
}

func (r *messageRepository) SoftDeleteByIDs(ctx context.Context, ids []string, deletedBy string) error {
	deletedByID, err := utils.ParseUUID(deletedBy, "deleted_by user ID")
	if err != nil {
		return err
	}

	now := time.Now()
	client := transaction.ResolveClient(ctx, r.client)

	for _, id := range ids {
		messageID, err := utils.ParseUUID(id, "message ID")
		if err != nil {
			return err
		}

		err = client.Message.UpdateOneID(messageID).
			SetDeletedAt(now).
			SetDeletedBy(deletedByID).
			Exec(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *messageRepository) AddReaction(ctx context.Context, reaction *entity.MessageReaction) error {
	messageID, err := utils.ParseUUID(reaction.MessageID, "message ID")
	if err != nil {
		return err
	}

	userID, err := utils.ParseUUID(reaction.UserID, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	mr, err := client.MessageReaction.Create().
		SetMessageID(messageID).
		SetUserID(userID).
		SetEmoji(reaction.Emoji).
		Save(ctx)
	if err != nil {
		return err
	}

	// Load edges
	mr, err = client.MessageReaction.Query().
		Where(
			messagereaction.HasMessageWith(message.ID(messageID)),
			messagereaction.HasUserWith(user.ID(userID)),
			messagereaction.Emoji(reaction.Emoji),
		).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
		WithUser().
		Only(ctx)
	if err != nil {
		return err
	}

	*reaction = *utils.MessageReactionToEntity(mr)
	return nil
}

func (r *messageRepository) RemoveReaction(ctx context.Context, messageID, userID, emoji string) error {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	_, err = client.MessageReaction.Delete().
		Where(
			messagereaction.HasMessageWith(message.ID(mid)),
			messagereaction.HasUserWith(user.ID(uid)),
			messagereaction.Emoji(emoji),
		).
		Exec(ctx)

	return err
}

func (r *messageRepository) FindReactions(ctx context.Context, messageID string) ([]*entity.MessageReaction, error) {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	reactions, err := client.MessageReaction.Query().
		Where(messagereaction.HasMessageWith(message.ID(mid))).
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

	result := make([]*entity.MessageReaction, 0, len(reactions))
	for _, mr := range reactions {
		result = append(result, utils.MessageReactionToEntity(mr))
	}

	return result, nil
}

func (r *messageRepository) FindReactionsByMessageIDs(ctx context.Context, messageIDs []string) (map[string][]*entity.MessageReaction, error) {
	if len(messageIDs) == 0 {
		return make(map[string][]*entity.MessageReaction), nil
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
	reactions, err := client.MessageReaction.Query().
		Where(messagereaction.HasMessageWith(message.IDIn(parsedIDs...))).
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

	result := make(map[string][]*entity.MessageReaction)
	for _, reaction := range reactions {
		messageID := reaction.Edges.Message.ID.String()
		if result[messageID] == nil {
			result[messageID] = make([]*entity.MessageReaction, 0)
		}
		result[messageID] = append(result[messageID], utils.MessageReactionToEntity(reaction))
	}

	return result, nil
}

func (r *messageRepository) AddUserMention(ctx context.Context, mention *entity.MessageUserMention) error {
	messageID, err := utils.ParseUUID(mention.MessageID, "message ID")
	if err != nil {
		return err
	}

	userID, err := utils.ParseUUID(mention.UserID, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	_, err = client.MessageUserMention.Create().
		SetMessageID(messageID).
		SetUserID(userID).
		Save(ctx)

	return err
}

func (r *messageRepository) AddGroupMention(ctx context.Context, mention *entity.MessageGroupMention) error {
	messageID, err := utils.ParseUUID(mention.MessageID, "message ID")
	if err != nil {
		return err
	}

	groupID, err := utils.ParseUUID(mention.GroupID, "group ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	_, err = client.MessageGroupMention.Create().
		SetMessageID(messageID).
		SetGroupID(groupID).
		Save(ctx)

	return err
}

func (r *messageRepository) Search(ctx context.Context, workspaceID, query string, limit int) ([]*entity.Message, error) {
	// For now, returning empty implementation
	// This requires full-text search which is better handled with PostgreSQL's full-text search or external search engine
	return []*entity.Message{}, nil
}
