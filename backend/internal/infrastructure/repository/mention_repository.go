package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/example/chat/internal/domain"
	"github.com/example/chat/internal/infrastructure/db"
)

type messageUserMentionRepository struct {
	db *gorm.DB
}

func NewMessageUserMentionRepository(db *gorm.DB) domain.MessageUserMentionRepository {
	return &messageUserMentionRepository{db: db}
}

func (r *messageUserMentionRepository) FindByMessageID(messageID string) ([]*domain.MessageUserMention, error) {
	msgID, err := uuid.Parse(messageID)
	if err != nil {
		return nil, errors.New("invalid message ID format")
	}

	var dbMentions []db.MessageUserMention
	if err := r.db.Where("message_id = ?", msgID).Order("created_at asc").Find(&dbMentions).Error; err != nil {
		return nil, err
	}

	mentions := make([]*domain.MessageUserMention, len(dbMentions))
	for i, m := range dbMentions {
		mentions[i] = toMessageUserMentionDomain(&m)
	}

	return mentions, nil
}

func (r *messageUserMentionRepository) FindByMessageIDs(messageIDs []string) ([]*domain.MessageUserMention, error) {
	if len(messageIDs) == 0 {
		return []*domain.MessageUserMention{}, nil
	}

	msgIDs := make([]uuid.UUID, len(messageIDs))
	for i, id := range messageIDs {
		msgID, err := uuid.Parse(id)
		if err != nil {
			return nil, errors.New("invalid message ID format")
		}
		msgIDs[i] = msgID
	}

	var dbMentions []db.MessageUserMention
	if err := r.db.Where("message_id IN ?", msgIDs).Order("message_id, created_at asc").Find(&dbMentions).Error; err != nil {
		return nil, err
	}

	mentions := make([]*domain.MessageUserMention, len(dbMentions))
	for i, m := range dbMentions {
		mentions[i] = toMessageUserMentionDomain(&m)
	}

	return mentions, nil
}

func (r *messageUserMentionRepository) FindByUserID(userID string, limit int, since *time.Time) ([]*domain.MessageUserMention, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	query := r.db.Where("user_id = ?", uid)

	if since != nil {
		query = query.Where("created_at > ?", since)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	var dbMentions []db.MessageUserMention
	if err := query.Order("created_at desc").Find(&dbMentions).Error; err != nil {
		return nil, err
	}

	mentions := make([]*domain.MessageUserMention, len(dbMentions))
	for i, m := range dbMentions {
		mentions[i] = toMessageUserMentionDomain(&m)
	}

	return mentions, nil
}

func (r *messageUserMentionRepository) Create(mention *domain.MessageUserMention) error {
	messageID, err := uuid.Parse(mention.MessageID)
	if err != nil {
		return errors.New("invalid message ID format")
	}

	userID, err := uuid.Parse(mention.UserID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	dbMention := &db.MessageUserMention{
		MessageID: messageID,
		UserID:    userID,
		CreatedAt: mention.CreatedAt,
	}

	return r.db.Create(dbMention).Error
}

func (r *messageUserMentionRepository) DeleteByMessageID(messageID string) error {
	msgID, err := uuid.Parse(messageID)
	if err != nil {
		return errors.New("invalid message ID format")
	}

	return r.db.Delete(&db.MessageUserMention{}, "message_id = ?", msgID).Error
}

func toMessageUserMentionDomain(dbMention *db.MessageUserMention) *domain.MessageUserMention {
	return &domain.MessageUserMention{
		MessageID: dbMention.MessageID.String(),
		UserID:    dbMention.UserID.String(),
		CreatedAt: dbMention.CreatedAt,
	}
}

type messageGroupMentionRepository struct {
	db *gorm.DB
}

func NewMessageGroupMentionRepository(db *gorm.DB) domain.MessageGroupMentionRepository {
	return &messageGroupMentionRepository{db: db}
}

func (r *messageGroupMentionRepository) FindByMessageID(messageID string) ([]*domain.MessageGroupMention, error) {
	msgID, err := uuid.Parse(messageID)
	if err != nil {
		return nil, errors.New("invalid message ID format")
	}

	var dbMentions []db.MessageGroupMention
	if err := r.db.Where("message_id = ?", msgID).Order("created_at asc").Find(&dbMentions).Error; err != nil {
		return nil, err
	}

	mentions := make([]*domain.MessageGroupMention, len(dbMentions))
	for i, m := range dbMentions {
		mentions[i] = toMessageGroupMentionDomain(&m)
	}

	return mentions, nil
}

func (r *messageGroupMentionRepository) FindByMessageIDs(messageIDs []string) ([]*domain.MessageGroupMention, error) {
	if len(messageIDs) == 0 {
		return []*domain.MessageGroupMention{}, nil
	}

	msgIDs := make([]uuid.UUID, len(messageIDs))
	for i, id := range messageIDs {
		msgID, err := uuid.Parse(id)
		if err != nil {
			return nil, errors.New("invalid message ID format")
		}
		msgIDs[i] = msgID
	}

	var dbMentions []db.MessageGroupMention
	if err := r.db.Where("message_id IN ?", msgIDs).Order("message_id, created_at asc").Find(&dbMentions).Error; err != nil {
		return nil, err
	}

	mentions := make([]*domain.MessageGroupMention, len(dbMentions))
	for i, m := range dbMentions {
		mentions[i] = toMessageGroupMentionDomain(&m)
	}

	return mentions, nil
}

func (r *messageGroupMentionRepository) FindByGroupID(groupID string, limit int, since *time.Time) ([]*domain.MessageGroupMention, error) {
	gID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, errors.New("invalid group ID format")
	}

	query := r.db.Where("group_id = ?", gID)

	if since != nil {
		query = query.Where("created_at > ?", since)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	var dbMentions []db.MessageGroupMention
	if err := query.Order("created_at desc").Find(&dbMentions).Error; err != nil {
		return nil, err
	}

	mentions := make([]*domain.MessageGroupMention, len(dbMentions))
	for i, m := range dbMentions {
		mentions[i] = toMessageGroupMentionDomain(&m)
	}

	return mentions, nil
}

func (r *messageGroupMentionRepository) Create(mention *domain.MessageGroupMention) error {
	messageID, err := uuid.Parse(mention.MessageID)
	if err != nil {
		return errors.New("invalid message ID format")
	}

	groupID, err := uuid.Parse(mention.GroupID)
	if err != nil {
		return errors.New("invalid group ID format")
	}

	dbMention := &db.MessageGroupMention{
		MessageID: messageID,
		GroupID:   groupID,
		CreatedAt: mention.CreatedAt,
	}

	return r.db.Create(dbMention).Error
}

func (r *messageGroupMentionRepository) DeleteByMessageID(messageID string) error {
	msgID, err := uuid.Parse(messageID)
	if err != nil {
		return errors.New("invalid message ID format")
	}

	return r.db.Delete(&db.MessageGroupMention{}, "message_id = ?", msgID).Error
}

func toMessageGroupMentionDomain(dbMention *db.MessageGroupMention) *domain.MessageGroupMention {
	return &domain.MessageGroupMention{
		MessageID: dbMention.MessageID.String(),
		GroupID:   dbMention.GroupID.String(),
		CreatedAt: dbMention.CreatedAt,
	}
}
