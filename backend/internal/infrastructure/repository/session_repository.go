package repository

import (
	"context"
	"time"

	"github.com/lib/pq"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/session"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type sessionRepository struct {
	client *ent.Client
}

func NewSessionRepository(client *ent.Client) domainrepository.SessionRepository {
	return &sessionRepository{client: client}
}

func (r *sessionRepository) FindByID(ctx context.Context, id string) (*entity.Session, error) {
	sessionID, err := parseUUID(id, "session ID")
	if err != nil {
		return nil, err
	}
	s, err := orNil(transaction.ResolveClient(ctx, r.client).Session.Get(ctx, sessionID))
	if s == nil {
		return nil, err
	}
	return sessionToEntity(s), nil
}

func (r *sessionRepository) FindActiveByTokenHash(ctx context.Context, refreshTokenHash string) (*entity.Session, error) {
	s, err := orNil(transaction.ResolveClient(ctx, r.client).Session.Query().
		Where(session.RefreshTokenHash(refreshTokenHash), session.ExpiresAtGT(time.Now()), session.RevokedAtIsNil()).
		Only(ctx))
	if s == nil {
		return nil, err
	}
	return sessionToEntity(s), nil
}

func (r *sessionRepository) Create(ctx context.Context, sess *entity.Session) error {
	uid, err := parseUUID(sess.UserID, "user ID")
	if err != nil {
		return err
	}
	s, err := transaction.ResolveClient(ctx, r.client).Session.Create().
		SetNillableID(parseUUIDPtr(&sess.ID)).
		SetUserID(uid).
		SetRefreshTokenHash(sess.RefreshTokenHash).
		SetExpiresAt(sess.ExpiresAt).
		SetIPAddress(sess.IPAddress).
		SetUserAgent(sess.UserAgent).
		Save(ctx)
	if err != nil {
		return err
	}
	*sess = *sessionToEntity(s)
	return nil
}

// Revoke は失効済みや存在しないセッションでもエラーにしません
func (r *sessionRepository) Revoke(ctx context.Context, id string) error {
	sessionID, err := parseUUID(id, "session ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).Session.Update().
		Where(session.ID(sessionID), session.RevokedAtIsNil()).
		SetRevokedAt(time.Now()).
		Exec(ctx)
}

func (r *sessionRepository) Rotate(ctx context.Context, id string, refreshTokenHash string, expiresAt time.Time) error {
	sessionID, err := parseUUID(id, "session ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).Session.UpdateOneID(sessionID).
		SetRefreshTokenHash(refreshTokenHash).
		SetExpiresAt(expiresAt).
		Exec(ctx)
}

// ユーザーごとに最後に作られたセッションを 1 件ずつ選ぶ。$1: ユーザー ID の配列
const latestSessionIDsSQL = `
	SELECT DISTINCT ON (user_id) id FROM session
	WHERE user_id = ANY($1::uuid[])
	ORDER BY user_id, created_at DESC`

func (r *sessionRepository) FindLatestByUserIDs(ctx context.Context, userIDs []string) (map[string]*entity.Session, error) {
	uids, err := parseUUIDs(userIDs, "user ID")
	if err != nil {
		return nil, err
	}
	ids, err := queryUUIDs(ctx, r.client, latestSessionIDsSQL, pq.Array(uids))
	if err != nil {
		return nil, err
	}
	sessions, err := transaction.ResolveClient(ctx, r.client).Session.Query().Where(session.IDIn(ids...)).All(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*entity.Session, len(sessions))
	for _, s := range sessions {
		result[s.UserID.String()] = sessionToEntity(s)
	}
	return result, nil
}

func (r *sessionRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).Session.Update().
		Where(session.UserID(uid), session.RevokedAtIsNil()).
		SetRevokedAt(time.Now()).
		Exec(ctx)
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	_, err := transaction.ResolveClient(ctx, r.client).Session.Delete().Where(session.ExpiresAtLT(time.Now())).Exec(ctx)
	return err
}
