package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/pushtoken"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type pushTokenRepository struct {
	client *ent.Client
}

func NewPushTokenRepository(client *ent.Client) domainrepository.PushTokenRepository {
	return &pushTokenRepository{client: client}
}

func (r *pushTokenRepository) Upsert(ctx context.Context, token *entity.PushToken) error {
	uid, err := parseUUID(token.UserID, "user ID")
	if err != nil {
		return err
	}

	return transaction.ResolveClient(ctx, r.client).PushToken.Create().
		SetUserID(uid).
		SetToken(token.Token).
		SetPlatform(pushtoken.Platform(token.Platform)).
		SetUserAgent(token.UserAgent).
		SetLastSeenAt(token.LastSeenAt).
		OnConflictColumns(pushtoken.FieldToken).
		UpdateNewValues().
		Exec(ctx)
}

func (r *pushTokenRepository) Delete(ctx context.Context, userID string, token string) error {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return err
	}

	_, err = transaction.ResolveClient(ctx, r.client).PushToken.Delete().
		Where(pushtoken.Token(token), pushtoken.UserID(uid)).
		Exec(ctx)
	return err
}

func (r *pushTokenRepository) DeleteTokens(ctx context.Context, tokens []string) error {
	if len(tokens) == 0 {
		return nil
	}
	_, err := transaction.ResolveClient(ctx, r.client).PushToken.Delete().Where(pushtoken.TokenIn(tokens...)).Exec(ctx)
	return err
}

func (r *pushTokenRepository) FindByUserIDs(ctx context.Context, userIDs []string) ([]*entity.PushToken, error) {
	uids, err := parseUUIDs(userIDs, "user ID")
	if err != nil {
		return nil, err
	}

	rows, err := transaction.ResolveClient(ctx, r.client).PushToken.Query().Where(pushtoken.UserIDIn(uids...)).All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(rows, func(row *ent.PushToken) *entity.PushToken {
		return &entity.PushToken{
			UserID:     row.UserID.String(),
			Token:      row.Token,
			Platform:   entity.PushPlatform(row.Platform),
			UserAgent:  row.UserAgent,
			LastSeenAt: row.LastSeenAt,
		}
	}), nil
}
