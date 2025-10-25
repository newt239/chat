package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/example/chat/internal/domain/entity"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"github.com/example/chat/internal/infrastructure/database"
)

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) domainrepository.MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) FindByID(ctx context.Context, id string) (*entity.Message, error) {
	messageID, err := parseUUID(id, "message ID")
	if err != nil {
		return nil, err
	}

	var model database.Message
	if err := r.db.WithContext(ctx).Where("id = ?", messageID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToEntity(), nil
}

func (r *messageRepository) FindByChannelID(ctx context.Context, channelID string, limit int, since *time.Time, until *time.Time) ([]*entity.Message, error) {
	chID, err := parseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}

	query := r.db.WithContext(ctx).Where("channel_id = ? AND parent_id IS NULL AND deleted_at IS NULL", chID)

	if since != nil {
		query = query.Where("created_at > ?", since)
	}

	if until != nil {
		query = query.Where("created_at < ?", until)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	var models []database.Message
	if err := query.Order("created_at desc").Find(&models).Error; err != nil {
		return nil, err
	}

	messages := make([]*entity.Message, len(models))
	for i, model := range models {
		messages[i] = model.ToEntity()
	}

	return messages, nil
}

func (r *messageRepository) FindThreadReplies(ctx context.Context, parentID string) ([]*entity.Message, error) {
	pID, err := parseUUID(parentID, "parent ID")
	if err != nil {
		return nil, err
	}

	var models []database.Message
	if err := r.db.WithContext(ctx).
		Where("parent_id = ? AND deleted_at IS NULL", pID).
		Order("created_at asc").
		Find(&models).Error; err != nil {
		return nil, err
	}

	messages := make([]*entity.Message, len(models))
	for i, model := range models {
		messages[i] = model.ToEntity()
	}

	return messages, nil
}

func (r *messageRepository) Create(ctx context.Context, message *entity.Message) error {
	channelID, err := parseUUID(message.ChannelID, "channel ID")
	if err != nil {
		return err
	}

	userID, err := parseUUID(message.UserID, "user ID")
	if err != nil {
		return err
	}

	model := &database.Message{}
	model.FromEntity(message)
	model.ChannelID = channelID
	model.UserID = userID

	if message.ID != "" {
		messageID, err := parseUUID(message.ID, "message ID")
		if err != nil {
			return err
		}
		model.ID = messageID
	}

	if message.ParentID != nil {
		parentID, err := parseUUID(*message.ParentID, "parent ID")
		if err != nil {
			return err
		}
		model.ParentID = &parentID
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	*message = *model.ToEntity()
	return nil
}

func (r *messageRepository) Update(ctx context.Context, message *entity.Message) error {
	messageID, err := parseUUID(message.ID, "message ID")
	if err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"body":      message.Body,
		"edited_at": now,
	}

	if err := r.db.WithContext(ctx).Model(&database.Message{}).Where("id = ?", messageID).Updates(updates).Error; err != nil {
		return err
	}

	message.EditedAt = &now

	return nil
}

func (r *messageRepository) Delete(ctx context.Context, id string) error {
	messageID, err := parseUUID(id, "message ID")
	if err != nil {
		return err
	}

	now := time.Now()
	return r.db.WithContext(ctx).Model(&database.Message{}).
		Where("id = ?", messageID).
		Update("deleted_at", now).Error
}

func (r *messageRepository) FindByChannelIDIncludingDeleted(ctx context.Context, channelID string, limit int, since *time.Time, until *time.Time) ([]*entity.Message, error) {
	chID, err := parseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}

	query := r.db.WithContext(ctx).Where("channel_id = ? AND parent_id IS NULL", chID)

	if since != nil {
		query = query.Where("created_at > ?", since)
	}

	if until != nil {
		query = query.Where("created_at < ?", until)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	var models []database.Message
	if err := query.Order("created_at desc").Find(&models).Error; err != nil {
		return nil, err
	}

	messages := make([]*entity.Message, len(models))
	for i, model := range models {
		messages[i] = model.ToEntity()
	}

	return messages, nil
}

func (r *messageRepository) FindThreadRepliesIncludingDeleted(ctx context.Context, parentID string) ([]*entity.Message, error) {
	pID, err := parseUUID(parentID, "parent ID")
	if err != nil {
		return nil, err
	}

	var models []database.Message
	if err := r.db.WithContext(ctx).
		Where("parent_id = ?", pID).
		Order("created_at asc").
		Find(&models).Error; err != nil {
		return nil, err
	}

	messages := make([]*entity.Message, len(models))
	for i, model := range models {
		messages[i] = model.ToEntity()
	}

	return messages, nil
}

func (r *messageRepository) SoftDeleteByIDs(ctx context.Context, ids []string, deletedBy string) error {
	if len(ids) == 0 {
		return nil
	}

	uuids := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		msgID, err := parseUUID(id, "message ID")
		if err != nil {
			return err
		}
		uuids = append(uuids, msgID)
	}

	deletedByUUID, err := parseUUID(deletedBy, "deleted by user ID")
	if err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"deleted_at": now,
		"deleted_by": deletedByUUID,
	}

	return r.db.WithContext(ctx).Model(&database.Message{}).
		Where("id IN ?", uuids).
		Updates(updates).Error
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

	model := &database.MessageReaction{}
	model.FromEntity(reaction)
	model.MessageID = messageID
	model.UserID = userID

	return r.db.WithContext(ctx).Create(model).Error
}

func (r *messageRepository) RemoveReaction(ctx context.Context, messageID string, userID string, emoji string) error {
	msgID, err := parseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&database.MessageReaction{}, "message_id = ? AND user_id = ? AND emoji = ?", msgID, uid, emoji).Error
}

func (r *messageRepository) FindReactions(ctx context.Context, messageID string) ([]*entity.MessageReaction, error) {
	msgID, err := parseUUID(messageID, "message ID")
	if err != nil {
		return nil, err
	}

	var models []database.MessageReaction
	if err := r.db.WithContext(ctx).Where("message_id = ?", msgID).Order("created_at asc").Find(&models).Error; err != nil {
		return nil, err
	}

	reactions := make([]*entity.MessageReaction, len(models))
	for i, model := range models {
		reactions[i] = model.ToEntity()
	}

	return reactions, nil
}

func (r *messageRepository) FindReactionsByMessageIDs(ctx context.Context, messageIDs []string) (map[string][]*entity.MessageReaction, error) {
	if len(messageIDs) == 0 {
		return make(map[string][]*entity.MessageReaction), nil
	}

	uuids := make([]uuid.UUID, 0, len(messageIDs))
	for _, id := range messageIDs {
		msgID, err := parseUUID(id, "message ID")
		if err != nil {
			return nil, err
		}
		uuids = append(uuids, msgID)
	}

	var models []database.MessageReaction
	if err := r.db.WithContext(ctx).Where("message_id IN ?", uuids).Order("created_at asc").Find(&models).Error; err != nil {
		return nil, err
	}

	result := make(map[string][]*entity.MessageReaction)
	for _, model := range models {
		messageID := model.MessageID.String()
		result[messageID] = append(result[messageID], model.ToEntity())
	}

	return result, nil
}
