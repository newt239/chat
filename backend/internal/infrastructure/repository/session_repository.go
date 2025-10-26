package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/models"
)

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) domainrepository.SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) FindByID(ctx context.Context, id string) (*entity.Session, error) {
	sessionID, err := parseUUID(id, "session ID")
	if err != nil {
		return nil, err
	}

	var model models.Session
	if err := r.db.WithContext(ctx).Where("id = ?", sessionID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToEntity(), nil
}

func (r *sessionRepository) FindActiveByUserID(ctx context.Context, userID string) ([]*entity.Session, error) {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	var models []models.Session
	now := time.Now()

	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", uid, now).
		Order("created_at desc").
		Find(&models).Error; err != nil {
		return nil, err
	}

	sessions := make([]*entity.Session, len(models))
	for i, model := range models {
		sessions[i] = model.ToEntity()
	}

	return sessions, nil
}

func (r *sessionRepository) Create(ctx context.Context, session *entity.Session) error {
	userID, err := parseUUID(session.UserID, "user ID")
	if err != nil {
		return err
	}

	model := &models.Session{}
	model.FromEntity(session)
	model.UserID = userID

	if session.ID != "" {
		sessionID, err := parseUUID(session.ID, "session ID")
		if err != nil {
			return err
		}
		model.ID = sessionID
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	*session = *model.ToEntity()
	return nil
}

func (r *sessionRepository) Revoke(ctx context.Context, id string) error {
	sessionID, err := parseUUID(id, "session ID")
	if err != nil {
		return err
	}

	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.Session{}).
		Where("id = ?", sessionID).
		Update("revoked_at", now).Error
}

func (r *sessionRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return err
	}

	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.Session{}).
		Where("user_id = ? AND revoked_at IS NULL", uid).
		Update("revoked_at", now).Error
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	return r.db.WithContext(ctx).Where("expires_at < ?", now).Delete(&models.Session{}).Error
}
