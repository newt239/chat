package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/example/chat/internal/domain"
	"github.com/example/chat/internal/infrastructure/db"
)

type messageLinkRepository struct {
	db *gorm.DB
}

func NewMessageLinkRepository(db *gorm.DB) domain.MessageLinkRepository {
	return &messageLinkRepository{db: db}
}

func (r *messageLinkRepository) FindByMessageID(messageID string) ([]*domain.MessageLink, error) {
	msgID, err := uuid.Parse(messageID)
	if err != nil {
		return nil, errors.New("invalid message ID format")
	}

	var dbLinks []db.MessageLink
	if err := r.db.Where("message_id = ?", msgID).Order("created_at asc").Find(&dbLinks).Error; err != nil {
		return nil, err
	}

	links := make([]*domain.MessageLink, len(dbLinks))
	for i, l := range dbLinks {
		links[i] = toMessageLinkDomain(&l)
	}

	return links, nil
}

func (r *messageLinkRepository) FindByMessageIDs(messageIDs []string) ([]*domain.MessageLink, error) {
	if len(messageIDs) == 0 {
		return []*domain.MessageLink{}, nil
	}

	msgIDs := make([]uuid.UUID, len(messageIDs))
	for i, id := range messageIDs {
		msgID, err := uuid.Parse(id)
		if err != nil {
			return nil, errors.New("invalid message ID format")
		}
		msgIDs[i] = msgID
	}

	var dbLinks []db.MessageLink
	if err := r.db.Where("message_id IN ?", msgIDs).Order("message_id, created_at asc").Find(&dbLinks).Error; err != nil {
		return nil, err
	}

	links := make([]*domain.MessageLink, len(dbLinks))
	for i, l := range dbLinks {
		links[i] = toMessageLinkDomain(&l)
	}

	return links, nil
}

func (r *messageLinkRepository) FindByURL(url string) (*domain.MessageLink, error) {
	var dbLink db.MessageLink
	if err := r.db.Where("url = ?", url).First(&dbLink).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toMessageLinkDomain(&dbLink), nil
}

func (r *messageLinkRepository) Create(link *domain.MessageLink) error {
	messageID, err := uuid.Parse(link.MessageID)
	if err != nil {
		return errors.New("invalid message ID format")
	}

	dbLink := &db.MessageLink{
		MessageID:   messageID,
		URL:         link.URL,
		Title:       link.Title,
		Description: link.Description,
		ImageURL:    link.ImageURL,
		SiteName:    link.SiteName,
		CardType:    link.CardType,
		CreatedAt:   link.CreatedAt,
	}

	if link.ID != "" {
		linkID, err := uuid.Parse(link.ID)
		if err != nil {
			return errors.New("invalid link ID format")
		}
		dbLink.ID = linkID
	}

	if err := r.db.Create(dbLink).Error; err != nil {
		return err
	}

	link.ID = dbLink.ID.String()

	return nil
}

func (r *messageLinkRepository) DeleteByMessageID(messageID string) error {
	msgID, err := uuid.Parse(messageID)
	if err != nil {
		return errors.New("invalid message ID format")
	}

	return r.db.Delete(&db.MessageLink{}, "message_id = ?", msgID).Error
}

func toMessageLinkDomain(dbLink *db.MessageLink) *domain.MessageLink {
	return &domain.MessageLink{
		ID:          dbLink.ID.String(),
		MessageID:   dbLink.MessageID.String(),
		URL:         dbLink.URL,
		Title:       dbLink.Title,
		Description: dbLink.Description,
		ImageURL:    dbLink.ImageURL,
		SiteName:    dbLink.SiteName,
		CardType:    dbLink.CardType,
		CreatedAt:   dbLink.CreatedAt,
	}
}
