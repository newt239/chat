package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/pushtoken"
	"github.com/newt239/chat/ent/user"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type pushTokenRepository struct {
	client *ent.Client
}

func NewPushTokenRepository(client *ent.Client) domainrepository.PushTokenRepository {
	return &pushTokenRepository{client: client}
}

func (r *pushTokenRepository) Upsert(ctx context.Context, token *entity.PushToken) error {
	uid, err := utils.ParseUUID(token.UserID, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	return client.PushToken.Create().
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
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	_, err = client.PushToken.Delete().
		Where(pushtoken.Token(token), pushtoken.HasUserWith(user.ID(uid))).
		Exec(ctx)
	return err
}

func (r *pushTokenRepository) DeleteTokens(ctx context.Context, tokens []string) error {
	if len(tokens) == 0 {
		return nil
	}
	client := transaction.ResolveClient(ctx, r.client)
	_, err := client.PushToken.Delete().Where(pushtoken.TokenIn(tokens...)).Exec(ctx)
	return err
}

func (r *pushTokenRepository) FindByUserIDs(ctx context.Context, userIDs []string) ([]*entity.PushToken, error) {
	uids, err := utils.ParseUUIDs(userIDs, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	rows, err := client.PushToken.Query().
		Where(pushtoken.HasUserWith(user.IDIn(uids...))).
		WithUser(func(q *ent.UserQuery) { q.Select(user.FieldID) }).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.PushToken, 0, len(rows))
	for _, row := range rows {
		result = append(result, &entity.PushToken{
			UserID:     row.Edges.User.ID.String(),
			Token:      row.Token,
			Platform:   entity.PushPlatform(row.Platform),
			UserAgent:  row.UserAgent,
			LastSeenAt: row.LastSeenAt,
		})
	}
	return result, nil
}
