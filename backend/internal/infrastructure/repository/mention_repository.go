package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/models"
)

type messageUserMentionRepository struct {
	db *gorm.DB
}

func NewMessageUserMentionRepository(db *gorm.DB) domainrepository.MessageUserMentionRepository {
	return &messageUserMentionRepository{db: db}
}

func (r *messageUserMentionRepository) FindByMessageID(ctx context.Context, messageID string) ([]*entity.MessageUserMention, error) {
	msgID, err := parseUUID(messageID, "message ID")
	if err != nil {
		return nil, err
	}

	var models []models.MessageUserMention
	if err := r.db.WithContext(ctx).Where("message_id = ?", msgID).Order("created_at asc").Find(&models).Error; err != nil {
		return nil, err
	}

	mentions := make([]*entity.MessageUserMention, len(models))
	for i, model := range models {
		mentions[i] = model.ToEntity()
	}

	return mentions, nil
}

func (r *messageUserMentionRepository) FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageUserMention, error) {
	if len(messageIDs) == 0 {
		return []*entity.MessageUserMention{}, nil
	}

	msgIDs := make([]uuid.UUID, len(messageIDs))
	for i, id := range messageIDs {
		msgID, err := parseUUID(id, "message ID")
		if err != nil {
			return nil, err
		}
		msgIDs[i] = msgID
	}

	var models []models.MessageUserMention
	if err := r.db.WithContext(ctx).Where("message_id IN ?", msgIDs).Order("message_id, created_at asc").Find(&models).Error; err != nil {
		return nil, err
	}

	mentions := make([]*entity.MessageUserMention, len(models))
	for i, model := range models {
		mentions[i] = model.ToEntity()
	}

	return mentions, nil
}

func (r *messageUserMentionRepository) FindByUserID(ctx context.Context, userID string, limit int, since *time.Time) ([]*entity.MessageUserMention, error) {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	query := r.db.WithContext(ctx).Where("user_id = ?", uid)

	if since != nil {
		query = query.Where("created_at > ?", since)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	var models []models.MessageUserMention
	if err := query.Order("created_at desc").Find(&models).Error; err != nil {
		return nil, err
	}

	mentions := make([]*entity.MessageUserMention, len(models))
	for i, model := range models {
		mentions[i] = model.ToEntity()
	}

	return mentions, nil
}

func (r *messageUserMentionRepository) Create(ctx context.Context, mention *entity.MessageUserMention) error {
	messageID, err := parseUUID(mention.MessageID, "message ID")
	if err != nil {
		return err
	}

	userID, err := parseUUID(mention.UserID, "user ID")
	if err != nil {
		return err
	}

	model := &models.MessageUserMention{}
	model.FromEntity(mention)
	model.MessageID = messageID
	model.UserID = userID

	return r.db.WithContext(ctx).Create(model).Error
}

func (r *messageUserMentionRepository) DeleteByMessageID(ctx context.Context, messageID string) error {
	msgID, err := parseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&models.MessageUserMention{}, "message_id = ?", msgID).Error
}

type messageGroupMentionRepository struct {
	db *gorm.DB
}

func NewMessageGroupMentionRepository(db *gorm.DB) domainrepository.MessageGroupMentionRepository {
	return &messageGroupMentionRepository{db: db}
}

func (r *messageGroupMentionRepository) FindByMessageID(ctx context.Context, messageID string) ([]*entity.MessageGroupMention, error) {
	msgID, err := parseUUID(messageID, "message ID")
	if err != nil {
		return nil, err
	}

	var models []models.MessageGroupMention
	if err := r.db.WithContext(ctx).Where("message_id = ?", msgID).Order("created_at asc").Find(&models).Error; err != nil {
		return nil, err
	}

	mentions := make([]*entity.MessageGroupMention, len(models))
	for i, model := range models {
		mentions[i] = model.ToEntity()
	}

	return mentions, nil
}

func (r *messageGroupMentionRepository) FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageGroupMention, error) {
	if len(messageIDs) == 0 {
		return []*entity.MessageGroupMention{}, nil
	}

	msgIDs := make([]uuid.UUID, len(messageIDs))
	for i, id := range messageIDs {
		msgID, err := parseUUID(id, "message ID")
		if err != nil {
			return nil, err
		}
		msgIDs[i] = msgID
	}

	var models []models.MessageGroupMention
	if err := r.db.WithContext(ctx).Where("message_id IN ?", msgIDs).Order("message_id, created_at asc").Find(&models).Error; err != nil {
		return nil, err
	}

	mentions := make([]*entity.MessageGroupMention, len(models))
	for i, model := range models {
		mentions[i] = model.ToEntity()
	}

	return mentions, nil
}

func (r *messageGroupMentionRepository) FindByGroupID(ctx context.Context, groupID string, limit int, since *time.Time) ([]*entity.MessageGroupMention, error) {
	gID, err := parseUUID(groupID, "group ID")
	if err != nil {
		return nil, err
	}

	query := r.db.WithContext(ctx).Where("group_id = ?", gID)

	if since != nil {
		query = query.Where("created_at > ?", since)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	var models []models.MessageGroupMention
	if err := query.Order("created_at desc").Find(&models).Error; err != nil {
		return nil, err
	}

	mentions := make([]*entity.MessageGroupMention, len(models))
	for i, model := range models {
		mentions[i] = model.ToEntity()
	}

	return mentions, nil
}

func (r *messageGroupMentionRepository) Create(ctx context.Context, mention *entity.MessageGroupMention) error {
	messageID, err := parseUUID(mention.MessageID, "message ID")
	if err != nil {
		return err
	}

	groupID, err := parseUUID(mention.GroupID, "group ID")
	if err != nil {
		return err
	}

	model := &models.MessageGroupMention{}
	model.FromEntity(mention)
	model.MessageID = messageID
	model.GroupID = groupID

	return r.db.WithContext(ctx).Create(model).Error
}

func (r *messageGroupMentionRepository) DeleteByMessageID(ctx context.Context, messageID string) error {
	msgID, err := parseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&models.MessageGroupMention{}, "message_id = ?", msgID).Error
}
