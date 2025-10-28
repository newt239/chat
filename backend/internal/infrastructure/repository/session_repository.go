package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/session"
	"github.com/newt239/chat/ent/user"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type sessionRepository struct {
	client *ent.Client
}

func NewSessionRepository(client *ent.Client) domainrepository.SessionRepository {
	return &sessionRepository{client: client}
}

func (r *sessionRepository) FindByID(ctx context.Context, id string) (*entity.Session, error) {
	sessionID, err := utils.ParseUUID(id, "session ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	s, err := client.Session.Query().
		Where(session.ID(sessionID)).
		WithUser().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return utils.SessionToEntity(s), nil
}

func (r *sessionRepository) FindActiveByUserID(ctx context.Context, userID string) ([]*entity.Session, error) {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	now := time.Now()
	client := transaction.ResolveClient(ctx, r.client)

	sessions, err := client.Session.Query().
		Where(
			session.HasUserWith(user.ID(uid)),
			session.ExpiresAtGT(now),
			session.RevokedAtIsNil(),
		).
		WithUser().
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.Session, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, utils.SessionToEntity(s))
	}

	return result, nil
}

func (r *sessionRepository) Create(ctx context.Context, sess *entity.Session) error {
	uid, err := utils.ParseUUID(sess.UserID, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	builder := client.Session.Create().
		SetUserID(uid).
		SetRefreshTokenHash(sess.RefreshTokenHash).
		SetExpiresAt(sess.ExpiresAt)

	if sess.ID != "" {
		sessionID, err := utils.ParseUUID(sess.ID, "session ID")
		if err != nil {
			return err
		}
		builder = builder.SetID(sessionID)
	}

	if sess.RevokedAt != nil {
		builder = builder.SetRevokedAt(*sess.RevokedAt)
	}

	s, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	// Load user edge
	s, err = client.Session.Query().
		Where(session.ID(s.ID)).
		WithUser().
		Only(ctx)
	if err != nil {
		return err
	}

	*sess = *utils.SessionToEntity(s)
	return nil
}

func (r *sessionRepository) Revoke(ctx context.Context, id string) error {
	sessionID, err := utils.ParseUUID(id, "session ID")
	if err != nil {
		return err
	}

	now := time.Now()
	client := transaction.ResolveClient(ctx, r.client)

	return client.Session.UpdateOneID(sessionID).
		SetRevokedAt(now).
		Exec(ctx)
}

func (r *sessionRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}

	now := time.Now()
	client := transaction.ResolveClient(ctx, r.client)

	_, err = client.Session.Update().
		Where(session.HasUserWith(user.ID(uid))).
		SetRevokedAt(now).
		Save(ctx)

	return err
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	client := transaction.ResolveClient(ctx, r.client)

	_, err := client.Session.Delete().
		Where(session.ExpiresAtLT(now)).
		Exec(ctx)

	return err
}
