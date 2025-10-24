package persistence

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/example/chat/internal/domain/entity"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"github.com/example/chat/internal/infrastructure/database"
)

type attachmentRepository struct {
	db *gorm.DB
}

func NewAttachmentRepository(db *gorm.DB) domainrepository.AttachmentRepository {
	return &attachmentRepository{db: db}
}

func (r *attachmentRepository) FindByID(ctx context.Context, id string) (*entity.Attachment, error) {
	attachmentID, err := parseUUID(id, "attachment ID")
	if err != nil {
		return nil, err
	}

	var model database.Attachment
	if err := r.db.WithContext(ctx).Where("id = ?", attachmentID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToEntity(), nil
}

func (r *attachmentRepository) FindByMessageID(ctx context.Context, messageID string) ([]*entity.Attachment, error) {
	msgID, err := parseUUID(messageID, "message ID")
	if err != nil {
		return nil, err
	}

	var models []database.Attachment
	if err := r.db.WithContext(ctx).Where("message_id = ?", msgID).Order("created_at asc").Find(&models).Error; err != nil {
		return nil, err
	}

	attachments := make([]*entity.Attachment, len(models))
	for i, model := range models {
		attachments[i] = model.ToEntity()
	}

	return attachments, nil
}

func (r *attachmentRepository) Create(ctx context.Context, attachment *entity.Attachment) error {
	messageID, err := parseUUID(attachment.MessageID, "message ID")
	if err != nil {
		return err
	}

	model := &database.Attachment{}
	model.FromEntity(attachment)
	model.MessageID = messageID

	if attachment.ID != "" {
		attachmentID, err := parseUUID(attachment.ID, "attachment ID")
		if err != nil {
			return err
		}
		model.ID = attachmentID
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	*attachment = *model.ToEntity()
	return nil
}

func (r *attachmentRepository) Delete(ctx context.Context, id string) error {
	attachmentID, err := parseUUID(id, "attachment ID")
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&database.Attachment{}, "id = ?", attachmentID).Error
}
