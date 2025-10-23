package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/example/chat/internal/domain"
	"github.com/example/chat/internal/infrastructure/db"
)

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) domain.SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) FindByID(id string) (*domain.Session, error) {
	sessionID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid session ID format")
	}

	var dbSession db.Session
	if err := r.db.Where("id = ?", sessionID).First(&dbSession).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toSessionDomain(&dbSession), nil
}

func (r *sessionRepository) FindActiveByUserID(userID string) ([]*domain.Session, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	var dbSessions []db.Session
	now := time.Now()

	if err := r.db.Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", uid, now).
		Order("created_at desc").
		Find(&dbSessions).Error; err != nil {
		return nil, err
	}

	sessions := make([]*domain.Session, len(dbSessions))
	for i, s := range dbSessions {
		sessions[i] = toSessionDomain(&s)
	}

	return sessions, nil
}

func (r *sessionRepository) Create(session *domain.Session) error {
	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	dbSession := &db.Session{
		UserID:           userID,
		RefreshTokenHash: session.RefreshTokenHash,
		ExpiresAt:        session.ExpiresAt,
	}

	if session.ID != "" {
		sessionID, err := uuid.Parse(session.ID)
		if err != nil {
			return errors.New("invalid session ID format")
		}
		dbSession.ID = sessionID
	}

	if err := r.db.Create(dbSession).Error; err != nil {
		return err
	}

	// Update domain object with generated values
	session.ID = dbSession.ID.String()
	session.CreatedAt = dbSession.CreatedAt

	return nil
}

func (r *sessionRepository) Revoke(id string) error {
	sessionID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid session ID format")
	}

	now := time.Now()
	return r.db.Model(&db.Session{}).
		Where("id = ?", sessionID).
		Update("revoked_at", now).Error
}

func (r *sessionRepository) RevokeAllByUserID(userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	now := time.Now()
	return r.db.Model(&db.Session{}).
		Where("user_id = ? AND revoked_at IS NULL", uid).
		Update("revoked_at", now).Error
}

func (r *sessionRepository) DeleteExpired() error {
	now := time.Now()
	return r.db.Where("expires_at < ?", now).Delete(&db.Session{}).Error
}

func toSessionDomain(dbSession *db.Session) *domain.Session {
	return &domain.Session{
		ID:               dbSession.ID.String(),
		UserID:           dbSession.UserID.String(),
		RefreshTokenHash: dbSession.RefreshTokenHash,
		ExpiresAt:        dbSession.ExpiresAt,
		RevokedAt:        dbSession.RevokedAt,
		CreatedAt:        dbSession.CreatedAt,
	}
}
