package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/message"
	"github.com/newt239/chat/ent/messagereaction"
	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type messageRepository struct {
	client *ent.Client
}

func NewMessageRepository(client *ent.Client) domainrepository.MessageRepository {
	return &messageRepository{client: client}
}

func (r *messageRepository) query(ctx context.Context) *ent.MessageQuery {
	return transaction.ResolveClient(ctx, r.client).Message.Query()
}

func (r *messageRepository) FindByID(ctx context.Context, id string) (*entity.Message, error) {
	messageID, err := parseUUID(id, "message ID")
	if err != nil {
		return nil, err
	}
	m, err := orNil(transaction.ResolveClient(ctx, r.client).Message.Get(ctx, messageID))
	if m == nil {
		return nil, err
	}
	return messageToEntity(m), nil
}

func (r *messageRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.Message, error) {
	parsedIDs, err := parseUUIDs(ids, "message ID")
	if err != nil {
		return nil, err
	}
	messages, err := r.query(ctx).Where(message.IDIn(parsedIDs...)).All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(messages, messageToEntity), nil
}

func (r *messageRepository) FindByChannelIDs(ctx context.Context, channelIDs []string, limit int, since *time.Time, until *time.Time, ascending bool) ([]*entity.Message, error) {
	cids, err := parseUUIDs(channelIDs, "channel ID")
	if err != nil {
		return nil, err
	}
	return r.page(ctx, r.query(ctx).Where(message.ChannelIDIn(cids...), message.ParentIDIsNil()), limit, since, until, ascending)
}

func (r *messageRepository) FindThreadReplies(ctx context.Context, parentID string, limit int, since *time.Time, until *time.Time, ascending bool) ([]*entity.Message, error) {
	pid, err := parseUUID(parentID, "parent ID")
	if err != nil {
		return nil, err
	}
	return r.page(ctx, r.query(ctx).Where(message.ParentID(pid)), limit, since, until, ascending)
}

// page は削除されていないメッセージを since より後・until より前で投稿日時の順に最大 limit 件返します。limit が 0 なら全件
func (r *messageRepository) page(ctx context.Context, query *ent.MessageQuery, limit int, since, until *time.Time, ascending bool) ([]*entity.Message, error) {
	query.Where(message.DeletedAtIsNil())
	if since != nil {
		query.Where(message.CreatedAtGT(*since))
	}
	if until != nil {
		query.Where(message.CreatedAtLT(*until))
	}
	if limit > 0 {
		query.Limit(limit)
	}
	order := ent.Desc(message.FieldCreatedAt)
	if ascending {
		order = ent.Asc(message.FieldCreatedAt)
	}
	messages, err := query.Order(order).All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(messages, messageToEntity), nil
}

func (r *messageRepository) Create(ctx context.Context, msg *entity.Message) error {
	channelID, err := parseUUID(msg.ChannelID, "channel ID")
	if err != nil {
		return err
	}
	userID, err := parseUUID(msg.UserID, "user ID")
	if err != nil {
		return err
	}
	builder := transaction.ResolveClient(ctx, r.client).Message.Create().
		SetChannelID(channelID).
		SetUserID(userID).
		SetNillableParentID(parseUUIDPtr(msg.ParentID)).
		SetBody(msg.Body).
		SetMentionsChannel(msg.MentionsChannel).
		SetMentionsHere(msg.MentionsHere)
	if loc := msg.Location; loc != nil {
		builder.
			SetLocationLatitude(loc.Latitude).
			SetLocationLongitude(loc.Longitude).
			SetNillableLocationAccuracy(loc.AccuracyMeters).
			SetNillableLocationLabel(loc.Label)
	}
	saved, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	*msg = *messageToEntity(saved)
	return nil
}

func (r *messageRepository) Update(ctx context.Context, msg *entity.Message) error {
	messageID, err := parseUUID(msg.ID, "message ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).Message.UpdateOneID(messageID).
		SetBody(msg.Body).
		SetMentionsChannel(msg.MentionsChannel).
		SetMentionsHere(msg.MentionsHere).
		SetNillableEditedAt(msg.EditedAt).
		Exec(ctx)
}

func (r *messageRepository) SoftDeleteByIDs(ctx context.Context, ids []string, deletedBy string) error {
	deletedByID, err := parseUUID(deletedBy, "deleted_by user ID")
	if err != nil {
		return err
	}
	messageIDs, err := parseUUIDs(ids, "message ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).Message.Update().
		Where(message.IDIn(messageIDs...)).
		SetDeletedAt(time.Now()).
		SetDeletedBy(deletedByID).
		Exec(ctx)
}

func (r *messageRepository) AddReaction(ctx context.Context, reaction *entity.MessageReaction) error {
	messageID, err := parseUUID(reaction.MessageID, "message ID")
	if err != nil {
		return err
	}
	userID, err := parseUUID(reaction.UserID, "user ID")
	if err != nil {
		return err
	}
	saved, err := transaction.ResolveClient(ctx, r.client).MessageReaction.Create().
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
	mid, err := parseUUID(messageID, "message ID")
	if err != nil {
		return err
	}
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return err
	}
	_, err = transaction.ResolveClient(ctx, r.client).MessageReaction.Delete().
		Where(messagereaction.MessageID(mid), messagereaction.UserID(uid), messagereaction.Emoji(emoji)).
		Exec(ctx)
	return err
}

func (r *messageRepository) FindReactionsByMessageIDs(ctx context.Context, messageIDs []string) (map[string][]*entity.MessageReaction, error) {
	parsedIDs, err := parseUUIDs(messageIDs, "message ID")
	if err != nil {
		return nil, err
	}
	reactions, err := transaction.ResolveClient(ctx, r.client).MessageReaction.Query().
		Where(messagereaction.MessageIDIn(parsedIDs...)).
		Order(ent.Asc(messagereaction.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string][]*entity.MessageReaction)
	for _, reaction := range reactions {
		messageID := reaction.MessageID.String()
		result[messageID] = append(result[messageID], &entity.MessageReaction{
			MessageID: messageID,
			UserID:    reaction.UserID.String(),
			Emoji:     reaction.Emoji,
			CreatedAt: reaction.CreatedAt,
		})
	}
	return result, nil
}
