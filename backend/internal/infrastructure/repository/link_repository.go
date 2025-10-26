package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/models"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type messageLinkRepository struct {
	db *gorm.DB
}

func NewMessageLinkRepository(db *gorm.DB) domainrepository.MessageLinkRepository {
	return &messageLinkRepository{db: db}
}

func (r *messageLinkRepository) FindByMessageID(ctx context.Context, messageID string) ([]*entity.MessageLink, error) {
	msgID, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return nil, err
	}

	var models []models.MessageLink
	if err := r.db.WithContext(ctx).Where("message_id = ?", msgID).Order("created_at asc").Find(&models).Error; err != nil {
		return nil, err
	}

	links := make([]*entity.MessageLink, len(models))
	for i, model := range models {
		links[i] = model.ToEntity()
	}

	return links, nil
}

func (r *messageLinkRepository) FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageLink, error) {
	if len(messageIDs) == 0 {
		return []*entity.MessageLink{}, nil
	}

	msgIDs := make([]uuid.UUID, len(messageIDs))
	for i, id := range messageIDs {
		msgID, err := utils.ParseUUID(id, "message ID")
		if err != nil {
			return nil, err
		}
		msgIDs[i] = msgID
	}

	var models []models.MessageLink
	if err := r.db.WithContext(ctx).Where("message_id IN ?", msgIDs).Order("message_id, created_at asc").Find(&models).Error; err != nil {
		return nil, err
	}

	links := make([]*entity.MessageLink, len(models))
	for i, model := range models {
		links[i] = model.ToEntity()
	}

	return links, nil
}

func (r *messageLinkRepository) FindByURL(ctx context.Context, url string) (*entity.MessageLink, error) {
	var model models.MessageLink
	if err := r.db.WithContext(ctx).Where("url = ?", url).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToEntity(), nil
}

func (r *messageLinkRepository) Create(ctx context.Context, link *entity.MessageLink) error {
	messageID, err := utils.ParseUUID(link.MessageID, "message ID")
	if err != nil {
		return err
	}

	model := &models.MessageLink{}
	model.FromEntity(link)
	model.MessageID = messageID

	if link.ID != "" {
		linkID, err := utils.ParseUUID(link.ID, "link ID")
		if err != nil {
			return err
		}
		model.ID = linkID
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	link.ID = model.ID.String()

	return nil
}

func (r *messageLinkRepository) DeleteByMessageID(ctx context.Context, messageID string) error {
	msgID, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&models.MessageLink{}, "message_id = ?", msgID).Error
}
