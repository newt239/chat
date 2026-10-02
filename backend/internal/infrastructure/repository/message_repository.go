package repository

import (
	"context"
	stdsql "database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/message"
	"github.com/newt239/chat/ent/messagereaction"
	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
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
	m, err := client.Message.Get(ctx, messageID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return utils.MessageToEntity(m), nil
}

func (r *messageRepository) FindByChannelIDs(ctx context.Context, channelIDs []string, limit int, since *time.Time, until *time.Time, ascending bool) ([]*entity.Message, error) {
	chIDs, err := parseUUIDs(channelIDs, "channel ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	query := client.Message.Query().
		Where(
			message.ChannelIDIn(chIDs...),
			message.ParentIDIsNil(),
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

	order := ent.Desc(message.FieldCreatedAt)
	if ascending {
		order = ent.Asc(message.FieldCreatedAt)
	}
	messages, err := query.Order(order).All(ctx)
	if err != nil {
		return nil, err
	}
	return toMessageEntities(messages), nil
}

func (r *messageRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.Message, error) {
	parsedIDs, err := parseUUIDs(ids, "message ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	messages, err := client.Message.Query().
		Where(message.IDIn(parsedIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toMessageEntities(messages), nil
}

func (r *messageRepository) FindThreadReplies(ctx context.Context, parentID string, limit int, since *time.Time, until *time.Time, ascending bool) ([]*entity.Message, error) {
	pID, err := utils.ParseUUID(parentID, "parent ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	query := client.Message.Query().
		Where(
			message.ParentID(pID),
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
	order := ent.Desc(message.FieldCreatedAt)
	if ascending {
		order = ent.Asc(message.FieldCreatedAt)
	}
	messages, err := query.Order(order).All(ctx)
	if err != nil {
		return nil, err
	}
	return toMessageEntities(messages), nil
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
		SetBody(msg.Body).
		SetMentionsChannel(msg.MentionsChannel).
		SetMentionsHere(msg.MentionsHere)
	if loc := msg.Location; loc != nil {
		builder = builder.
			SetLocationLatitude(loc.Latitude).
			SetLocationLongitude(loc.Longitude).
			SetNillableLocationAccuracy(loc.AccuracyMeters).
			SetNillableLocationLabel(loc.Label)
	}

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

	saved, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	*msg = *utils.MessageToEntity(saved)
	return nil
}

func (r *messageRepository) Update(ctx context.Context, msg *entity.Message) error {
	messageID, err := utils.ParseUUID(msg.ID, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	builder := client.Message.UpdateOneID(messageID).
		SetBody(msg.Body).
		SetMentionsChannel(msg.MentionsChannel).
		SetMentionsHere(msg.MentionsHere)

	if msg.EditedAt != nil {
		builder = builder.SetEditedAt(*msg.EditedAt)
	}

	saved, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	if !saved.EditedAt.IsZero() {
		msg.EditedAt = &saved.EditedAt
	}
	return nil
}

func (r *messageRepository) SoftDeleteByIDs(ctx context.Context, ids []string, deletedBy string) error {
	deletedByID, err := utils.ParseUUID(deletedBy, "deleted_by user ID")
	if err != nil {
		return err
	}
	messageIDs, err := parseUUIDs(ids, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	return client.Message.Update().
		Where(message.IDIn(messageIDs...)).
		SetDeletedAt(time.Now()).
		SetDeletedBy(deletedByID).
		Exec(ctx)
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

	saved, err := client.MessageReaction.Create().
		SetMessageID(messageID).
		SetUserID(userID).
		SetEmoji(reaction.Emoji).
		Save(ctx)
	if ent.IsConstraintError(err) {
		return domerr.ErrReactionExists
	}
	if err != nil {
		return err
	}

	reaction.CreatedAt = saved.CreatedAt
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
			messagereaction.MessageID(mid),
			messagereaction.UserID(uid),
			messagereaction.Emoji(emoji),
		).
		Exec(ctx)

	return err
}

func (r *messageRepository) FindReactions(ctx context.Context, messageID string) ([]*entity.MessageReaction, error) {
	reactions, err := r.FindReactionsByMessageIDs(ctx, []string{messageID})
	if err != nil {
		return nil, err
	}
	return reactions[messageID], nil
}

func (r *messageRepository) FindReactionsByMessageIDs(ctx context.Context, messageIDs []string) (map[string][]*entity.MessageReaction, error) {
	result := make(map[string][]*entity.MessageReaction)
	if len(messageIDs) == 0 {
		return result, nil
	}

	parsedIDs, err := parseUUIDs(messageIDs, "message ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	reactions, err := client.MessageReaction.Query().
		Where(messagereaction.MessageIDIn(parsedIDs...)).
		Order(ent.Asc(messagereaction.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	for _, reaction := range reactions {
		messageID := reaction.MessageID.String()
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
		SetNillableViaGroupID(utils.ParseUUIDPtr(mention.ViaGroupID)).
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

// ignoreConflict は ON CONFLICT DO NOTHING で既存の行と重なったときに返る sql.ErrNoRows を無視します
func ignoreConflict(err error) error {
	if errors.Is(err, stdsql.ErrNoRows) {
		return nil
	}
	return err
}

func parseUUIDs(ids []string, label string) ([]uuid.UUID, error) {
	parsed := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		p, err := utils.ParseUUID(id, label)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, p)
	}
	return parsed, nil
}

func toMessageEntities(messages []*ent.Message) []*entity.Message {
	result := make([]*entity.Message, 0, len(messages))
	for _, m := range messages {
		result = append(result, utils.MessageToEntity(m))
	}
	return result
}
