package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/example/chat/internal/domain"
	"github.com/example/chat/internal/infrastructure/db"
)

type attachmentRepository struct {
	db *gorm.DB
}

func NewAttachmentRepository(db *gorm.DB) domain.AttachmentRepository {
	return &attachmentRepository{db: db}
}

func (r *attachmentRepository) FindByID(id string) (*domain.Attachment, error) {
	attachmentID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid attachment ID format")
	}

	var dbAttachment db.Attachment
	if err := r.db.Where("id = ?", attachmentID).First(&dbAttachment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toAttachmentDomain(&dbAttachment), nil
}

func (r *attachmentRepository) FindByMessageID(messageID string) ([]*domain.Attachment, error) {
	msgID, err := uuid.Parse(messageID)
	if err != nil {
		return nil, errors.New("invalid message ID format")
	}

	var dbAttachments []db.Attachment
	if err := r.db.Where("message_id = ?", msgID).Order("created_at asc").Find(&dbAttachments).Error; err != nil {
		return nil, err
	}

	attachments := make([]*domain.Attachment, len(dbAttachments))
	for i, a := range dbAttachments {
		attachments[i] = toAttachmentDomain(&a)
	}

	return attachments, nil
}

func (r *attachmentRepository) Create(attachment *domain.Attachment) error {
	messageID, err := uuid.Parse(attachment.MessageID)
	if err != nil {
		return errors.New("invalid message ID format")
	}

	dbAttachment := &db.Attachment{
		MessageID:  messageID,
		FileName:   attachment.FileName,
		MimeType:   attachment.MimeType,
		SizeBytes:  attachment.SizeBytes,
		StorageKey: attachment.StorageKey,
	}

	if attachment.ID != "" {
		attachmentID, err := uuid.Parse(attachment.ID)
		if err != nil {
			return errors.New("invalid attachment ID format")
		}
		dbAttachment.ID = attachmentID
	}

	if err := r.db.Create(dbAttachment).Error; err != nil {
		return err
	}

	attachment.ID = dbAttachment.ID.String()
	attachment.CreatedAt = dbAttachment.CreatedAt

	return nil
}

func (r *attachmentRepository) Delete(id string) error {
	attachmentID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid attachment ID format")
	}

	return r.db.Delete(&db.Attachment{}, "id = ?", attachmentID).Error
}

func toAttachmentDomain(dbAttachment *db.Attachment) *domain.Attachment {
	return &domain.Attachment{
		ID:         dbAttachment.ID.String(),
		MessageID:  dbAttachment.MessageID.String(),
		FileName:   dbAttachment.FileName,
		MimeType:   dbAttachment.MimeType,
		SizeBytes:  dbAttachment.SizeBytes,
		StorageKey: dbAttachment.StorageKey,
		CreatedAt:  dbAttachment.CreatedAt,
	}
}
