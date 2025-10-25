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

func (r *attachmentRepository) FindByMessageIDs(ctx context.Context, messageIDs []string) (map[string][]*entity.Attachment, error) {
	if len(messageIDs) == 0 {
		return make(map[string][]*entity.Attachment), nil
	}

	msgIDs := make([]interface{}, len(messageIDs))
	for i, id := range messageIDs {
		parsed, err := parseUUID(id, "message ID")
		if err != nil {
			return nil, err
		}
		msgIDs[i] = parsed
	}

	var models []database.Attachment
	if err := r.db.WithContext(ctx).Where("message_id IN ?", msgIDs).Order("created_at asc").Find(&models).Error; err != nil {
		return nil, err
	}

	result := make(map[string][]*entity.Attachment)
	for _, model := range models {
		entity := model.ToEntity()
		if entity.MessageID != nil {
			result[*entity.MessageID] = append(result[*entity.MessageID], entity)
		}
	}

	return result, nil
}

func (r *attachmentRepository) FindPendingByIDsForUser(ctx context.Context, userID string, attachmentIDs []string) ([]*entity.Attachment, error) {
	if len(attachmentIDs) == 0 {
		return []*entity.Attachment{}, nil
	}

	userUUID, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	ids := make([]interface{}, len(attachmentIDs))
	for i, id := range attachmentIDs {
		parsed, err := parseUUID(id, "attachment ID")
		if err != nil {
			return nil, err
		}
		ids[i] = parsed
	}

	var models []database.Attachment
	if err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Where("uploader_id = ?", userUUID).
		Where("status = ?", entity.AttachmentStatusPending).
		Find(&models).Error; err != nil {
		return nil, err
	}

	attachments := make([]*entity.Attachment, len(models))
	for i, model := range models {
		attachments[i] = model.ToEntity()
	}

	return attachments, nil
}

func (r *attachmentRepository) CreatePending(ctx context.Context, attachment *entity.Attachment) error {
	model := &database.Attachment{}
	model.FromEntity(attachment)

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

func (r *attachmentRepository) AttachToMessage(ctx context.Context, attachmentIDs []string, messageID string) error {
	if len(attachmentIDs) == 0 {
		return nil
	}

	msgID, err := parseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	ids := make([]interface{}, len(attachmentIDs))
	for i, id := range attachmentIDs {
		parsed, err := parseUUID(id, "attachment ID")
		if err != nil {
			return err
		}
		ids[i] = parsed
	}

	now := r.db.NowFunc()
	updates := map[string]interface{}{
		"message_id":  msgID,
		"status":      entity.AttachmentStatusAttached,
		"uploaded_at": now,
	}

	return r.db.WithContext(ctx).
		Model(&database.Attachment{}).
		Where("id IN ?", ids).
		Updates(updates).
		Error
}

func (r *attachmentRepository) Delete(ctx context.Context, id string) error {
	attachmentID, err := parseUUID(id, "attachment ID")
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&database.Attachment{}, "id = ?", attachmentID).Error
}
