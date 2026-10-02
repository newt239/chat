package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/session"
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
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return utils.SessionToEntity(s), nil
}

func (r *sessionRepository) FindActiveByTokenHash(ctx context.Context, refreshTokenHash string) (*entity.Session, error) {
	s, err := transaction.ResolveClient(ctx, r.client).Session.Query().
		Where(
			session.RefreshTokenHash(refreshTokenHash),
			session.ExpiresAtGT(time.Now()),
			session.RevokedAtIsNil(),
		).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return utils.SessionToEntity(s), nil
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
		SetExpiresAt(sess.ExpiresAt).
		SetIPAddress(sess.IPAddress).
		SetUserAgent(sess.UserAgent)

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
	*sess = *utils.SessionToEntity(s)
	return nil
}

// Revoke は失効済みや存在しないセッションでもエラーにしません
func (r *sessionRepository) Revoke(ctx context.Context, id string) error {
	sessionID, err := utils.ParseUUID(id, "session ID")
	if err != nil {
		return err
	}
	_, err = transaction.ResolveClient(ctx, r.client).Session.Update().
		Where(session.ID(sessionID), session.RevokedAtIsNil()).
		SetRevokedAt(time.Now()).
		Save(ctx)
	return err
}

func (r *sessionRepository) Rotate(ctx context.Context, id string, refreshTokenHash string, expiresAt time.Time) error {
	sessionID, err := utils.ParseUUID(id, "session ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	return client.Session.UpdateOneID(sessionID).
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
	result := make(map[string]*entity.Session, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}
	uids, err := utils.ParseUUIDs(userIDs, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	rows, err := client.QueryContext(ctx, latestSessionIDsSQL, pq.Array(uids))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sessions, err := client.Session.Query().Where(session.IDIn(ids...)).All(ctx)
	if err != nil {
		return nil, err
	}
	for _, s := range sessions {
		e := utils.SessionToEntity(s)
		result[e.UserID] = e
	}
	return result, nil
}

func (r *sessionRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}

	now := time.Now()
	client := transaction.ResolveClient(ctx, r.client)

	_, err = client.Session.Update().
		Where(session.UserID(uid), session.RevokedAtIsNil()).
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
