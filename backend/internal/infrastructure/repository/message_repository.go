package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/example/chat/internal/domain"
	"github.com/example/chat/internal/infrastructure/db"
)

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) domain.MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) FindByID(id string) (*domain.Message, error) {
	messageID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid message ID format")
	}

	var dbMessage db.Message
	if err := r.db.Where("id = ?", messageID).First(&dbMessage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toMessageDomain(&dbMessage), nil
}

func (r *messageRepository) FindByChannelID(channelID string, limit int, since, until *time.Time) ([]*domain.Message, error) {
	chID, err := uuid.Parse(channelID)
	if err != nil {
		return nil, errors.New("invalid channel ID format")
	}

	query := r.db.Where("channel_id = ? AND parent_id IS NULL AND deleted_at IS NULL", chID)

	if since != nil {
		query = query.Where("created_at > ?", since)
	}

	if until != nil {
		query = query.Where("created_at < ?", until)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	var dbMessages []db.Message
	if err := query.Order("created_at desc").Find(&dbMessages).Error; err != nil {
		return nil, err
	}

	messages := make([]*domain.Message, len(dbMessages))
	for i, m := range dbMessages {
		messages[i] = toMessageDomain(&m)
	}

	return messages, nil
}

func (r *messageRepository) FindThreadReplies(parentID string) ([]*domain.Message, error) {
	pID, err := uuid.Parse(parentID)
	if err != nil {
		return nil, errors.New("invalid parent ID format")
	}

	var dbMessages []db.Message
	if err := r.db.
		Where("parent_id = ? AND deleted_at IS NULL", pID).
		Order("created_at asc").
		Find(&dbMessages).Error; err != nil {
		return nil, err
	}

	messages := make([]*domain.Message, len(dbMessages))
	for i, m := range dbMessages {
		messages[i] = toMessageDomain(&m)
	}

	return messages, nil
}

func (r *messageRepository) Create(message *domain.Message) error {
	channelID, err := uuid.Parse(message.ChannelID)
	if err != nil {
		return errors.New("invalid channel ID format")
	}

	userID, err := uuid.Parse(message.UserID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	dbMessage := &db.Message{
		ChannelID: channelID,
		UserID:    userID,
		Body:      message.Body,
	}

	if message.ID != "" {
		messageID, err := uuid.Parse(message.ID)
		if err != nil {
			return errors.New("invalid message ID format")
		}
		dbMessage.ID = messageID
	}

	if message.ParentID != nil {
		parentID, err := uuid.Parse(*message.ParentID)
		if err != nil {
			return errors.New("invalid parent ID format")
		}
		dbMessage.ParentID = &parentID
	}

	if err := r.db.Create(dbMessage).Error; err != nil {
		return err
	}

	message.ID = dbMessage.ID.String()
	message.CreatedAt = dbMessage.CreatedAt

	return nil
}

func (r *messageRepository) Update(message *domain.Message) error {
	messageID, err := uuid.Parse(message.ID)
	if err != nil {
		return errors.New("invalid message ID format")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"body":      message.Body,
		"edited_at": now,
	}

	if err := r.db.Model(&db.Message{}).Where("id = ?", messageID).Updates(updates).Error; err != nil {
		return err
	}

	message.EditedAt = &now

	return nil
}

func (r *messageRepository) Delete(id string) error {
	messageID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid message ID format")
	}

	now := time.Now()
	return r.db.Model(&db.Message{}).
		Where("id = ?", messageID).
		Update("deleted_at", now).Error
}

func (r *messageRepository) AddReaction(reaction *domain.MessageReaction) error {
	messageID, err := uuid.Parse(reaction.MessageID)
	if err != nil {
		return errors.New("invalid message ID format")
	}

	userID, err := uuid.Parse(reaction.UserID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	dbReaction := &db.MessageReaction{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     reaction.Emoji,
	}

	return r.db.Create(dbReaction).Error
}

func (r *messageRepository) RemoveReaction(messageID, userID, emoji string) error {
	msgID, err := uuid.Parse(messageID)
	if err != nil {
		return errors.New("invalid message ID format")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	return r.db.Delete(&db.MessageReaction{}, "message_id = ? AND user_id = ? AND emoji = ?", msgID, uid, emoji).Error
}

func (r *messageRepository) FindReactions(messageID string) ([]*domain.MessageReaction, error) {
	msgID, err := uuid.Parse(messageID)
	if err != nil {
		return nil, errors.New("invalid message ID format")
	}

	var dbReactions []db.MessageReaction
	if err := r.db.Where("message_id = ?", msgID).Order("created_at asc").Find(&dbReactions).Error; err != nil {
		return nil, err
	}

	reactions := make([]*domain.MessageReaction, len(dbReactions))
	for i, r := range dbReactions {
		reactions[i] = toMessageReactionDomain(&r)
	}

	return reactions, nil
}

func toMessageDomain(dbMessage *db.Message) *domain.Message {
	message := &domain.Message{
		ID:        dbMessage.ID.String(),
		ChannelID: dbMessage.ChannelID.String(),
		UserID:    dbMessage.UserID.String(),
		Body:      dbMessage.Body,
		CreatedAt: dbMessage.CreatedAt,
		EditedAt:  dbMessage.EditedAt,
		DeletedAt: dbMessage.DeletedAt,
	}

	if dbMessage.ParentID != nil {
		parentID := dbMessage.ParentID.String()
		message.ParentID = &parentID
	}

	return message
}

func toMessageReactionDomain(dbReaction *db.MessageReaction) *domain.MessageReaction {
	return &domain.MessageReaction{
		MessageID: dbReaction.MessageID.String(),
		UserID:    dbReaction.UserID.String(),
		Emoji:     dbReaction.Emoji,
		CreatedAt: dbReaction.CreatedAt,
	}
}
