package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/webhook"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type webhookRepository struct {
	client *ent.Client
}

func NewWebhookRepository(client *ent.Client) domainrepository.WebhookRepository {
	return &webhookRepository{client: client}
}

func (r *webhookRepository) query(ctx context.Context) *ent.WebhookQuery {
	return transaction.ResolveClient(ctx, r.client).Webhook.Query().
		WithChannel().
		WithCreatedBy().
		WithBotUser()
}

func (r *webhookRepository) FindByID(ctx context.Context, id string) (*entity.Webhook, error) {
	webhookID, err := utils.ParseUUID(id, "webhook ID")
	if err != nil {
		return nil, err
	}
	found, err := r.query(ctx).Where(webhook.ID(webhookID)).Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return webhookToEntity(found), nil
}

func (r *webhookRepository) FindByChannelID(ctx context.Context, channelID string) ([]*entity.Webhook, error) {
	cid, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}
	found, err := r.query(ctx).
		Where(webhook.HasChannelWith(channel.ID(cid))).
		Order(ent.Asc(webhook.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*entity.Webhook, 0, len(found))
	for _, w := range found {
		result = append(result, webhookToEntity(w))
	}
	return result, nil
}

func (r *webhookRepository) Create(ctx context.Context, w *entity.Webhook) error {
	cid, err := utils.ParseUUID(w.ChannelID, "channel ID")
	if err != nil {
		return err
	}
	creatorID, err := utils.ParseUUID(w.CreatedBy, "user ID")
	if err != nil {
		return err
	}
	botID, err := utils.ParseUUID(w.BotUserID, "bot user ID")
	if err != nil {
		return err
	}

	created, err := transaction.ResolveClient(ctx, r.client).Webhook.Create().
		SetChannelID(cid).
		SetCreatedByID(creatorID).
		SetBotUserID(botID).
		SetName(w.Name).
		SetNillableAvatarURL(w.AvatarURL).
		SetTokenHash(w.TokenHash).
		Save(ctx)
	if err != nil {
		return err
	}
	w.ID = created.ID.String()
	w.CreatedAt = created.CreatedAt
	w.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *webhookRepository) Update(ctx context.Context, w *entity.Webhook) error {
	webhookID, err := utils.ParseUUID(w.ID, "webhook ID")
	if err != nil {
		return err
	}
	builder := transaction.ResolveClient(ctx, r.client).Webhook.UpdateOneID(webhookID).
		SetName(w.Name).
		SetTokenHash(w.TokenHash)
	if w.AvatarURL != nil {
		builder = builder.SetAvatarURL(*w.AvatarURL)
	} else {
		builder = builder.ClearAvatarURL()
	}
	updated, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	w.UpdatedAt = updated.UpdatedAt
	return nil
}

func (r *webhookRepository) MarkUsed(ctx context.Context, id string, usedAt time.Time) error {
	webhookID, err := utils.ParseUUID(id, "webhook ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).Webhook.UpdateOneID(webhookID).SetLastUsedAt(usedAt).Exec(ctx)
}

func (r *webhookRepository) Delete(ctx context.Context, id string) error {
	webhookID, err := utils.ParseUUID(id, "webhook ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).Webhook.DeleteOneID(webhookID).Exec(ctx)
}

func webhookToEntity(w *ent.Webhook) *entity.Webhook {
	result := &entity.Webhook{
		ID:         w.ID.String(),
		Name:       w.Name,
		AvatarURL:  w.AvatarURL,
		TokenHash:  w.TokenHash,
		LastUsedAt: w.LastUsedAt,
		CreatedAt:  w.CreatedAt,
		UpdatedAt:  w.UpdatedAt,
	}
	if w.Edges.Channel != nil {
		result.ChannelID = w.Edges.Channel.ID.String()
	}
	if w.Edges.CreatedBy != nil {
		result.CreatedBy = w.Edges.CreatedBy.ID.String()
	}
	if w.Edges.BotUser != nil {
		result.BotUserID = w.Edges.BotUser.ID.String()
	}
	return result
}
