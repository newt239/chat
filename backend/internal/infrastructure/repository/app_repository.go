package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/app"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/channelmember"
	"github.com/newt239/chat/ent/user"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type appRepository struct {
	client *ent.Client
}

func NewAppRepository(client *ent.Client) domainrepository.AppRepository {
	return &appRepository{client: client}
}

func (r *appRepository) query(ctx context.Context) *ent.AppQuery {
	return transaction.ResolveClient(ctx, r.client).App.Query().
		WithCreatedBy().
		WithBotUser().
		WithDefaultChannel()
}

func (r *appRepository) find(ctx context.Context, query *ent.AppQuery) ([]*entity.App, error) {
	found, err := query.Order(ent.Asc(app.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*entity.App, 0, len(found))
	for _, a := range found {
		result = append(result, appToEntity(a))
	}
	return result, nil
}

func (r *appRepository) findOne(ctx context.Context, query *ent.AppQuery) (*entity.App, error) {
	found, err := query.Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return appToEntity(found), nil
}

func (r *appRepository) FindByID(ctx context.Context, id string) (*entity.App, error) {
	appID, err := utils.ParseUUID(id, "app ID")
	if err != nil {
		return nil, err
	}
	return r.findOne(ctx, r.query(ctx).Where(app.ID(appID)))
}

func (r *appRepository) FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.App, error) {
	return r.find(ctx, r.query(ctx).Where(app.WorkspaceID(workspaceID)))
}

func (r *appRepository) FindOfficial(ctx context.Context, workspaceID string) (*entity.App, error) {
	return r.findOne(ctx, r.query(ctx).Where(app.WorkspaceID(workspaceID), app.IsOfficial(true)))
}

func (r *appRepository) FindByChannelID(ctx context.Context, channelID string) ([]*entity.App, error) {
	cid, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}
	return r.find(ctx, r.query(ctx).Where(app.HasBotUserWith(
		user.HasChannelMembersWith(channelmember.HasChannelWith(channel.ID(cid))),
	)))
}

func (r *appRepository) Create(ctx context.Context, a *entity.App) error {
	creatorID, err := utils.ParseUUID(a.CreatedBy, "user ID")
	if err != nil {
		return err
	}
	botID, err := utils.ParseUUID(a.BotUserID, "bot user ID")
	if err != nil {
		return err
	}
	builder := transaction.ResolveClient(ctx, r.client).App.Create().
		SetWorkspaceID(a.WorkspaceID).
		SetCreatedByID(creatorID).
		SetBotUserID(botID).
		SetIsOfficial(a.IsOfficial).
		SetName(a.Name).
		SetNillableDescription(a.Description).
		SetNillableAvatarURL(a.AvatarURL).
		SetNillableTokenHash(a.TokenHash).
		SetPermissions(permissionStrings(a.Permissions)).
		SetNillableDefaultChannelID(utils.ParseUUIDPtr(a.DefaultChannelID)).
		SetNillableOutgoingURL(a.OutgoingURL).
		SetNillableOutgoingSecret(a.OutgoingSecret)
	if a.ID != "" {
		appID, err := utils.ParseUUID(a.ID, "app ID")
		if err != nil {
			return err
		}
		builder = builder.SetID(appID)
	}
	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	a.ID = created.ID.String()
	a.CreatedAt = created.CreatedAt
	return nil
}

func (r *appRepository) Update(ctx context.Context, a *entity.App) error {
	appID, err := utils.ParseUUID(a.ID, "app ID")
	if err != nil {
		return err
	}
	builder := transaction.ResolveClient(ctx, r.client).App.UpdateOneID(appID).
		SetName(a.Name).
		SetPermissions(permissionStrings(a.Permissions))
	if a.Description != nil {
		builder.SetDescription(*a.Description)
	} else {
		builder.ClearDescription()
	}
	if a.AvatarURL != nil {
		builder.SetAvatarURL(*a.AvatarURL)
	} else {
		builder.ClearAvatarURL()
	}
	if a.TokenHash != nil {
		builder.SetTokenHash(*a.TokenHash)
	} else {
		builder.ClearTokenHash()
	}
	if id := utils.ParseUUIDPtr(a.DefaultChannelID); id != nil {
		builder.SetDefaultChannelID(*id)
	} else {
		builder.ClearDefaultChannel()
	}
	if a.OutgoingURL != nil {
		builder.SetOutgoingURL(*a.OutgoingURL)
	} else {
		builder.ClearOutgoingURL()
	}
	if a.OutgoingSecret != nil {
		builder.SetOutgoingSecret(*a.OutgoingSecret)
	} else {
		builder.ClearOutgoingSecret()
	}
	return builder.Exec(ctx)
}

func (r *appRepository) MarkUsed(ctx context.Context, id string, usedAt time.Time) error {
	appID, err := utils.ParseUUID(id, "app ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).App.UpdateOneID(appID).SetLastUsedAt(usedAt).Exec(ctx)
}

// Delete はアプリと、ボットユーザーのチャンネルへの参加を消します。ボットユーザーは過去の投稿の名義として残す
func (r *appRepository) Delete(ctx context.Context, id string) error {
	appID, err := utils.ParseUUID(id, "app ID")
	if err != nil {
		return err
	}
	client := transaction.ResolveClient(ctx, r.client)
	botID, err := client.App.Query().Where(app.ID(appID)).QueryBotUser().OnlyID(ctx)
	if err != nil {
		return err
	}
	if _, err := client.ChannelMember.Delete().Where(channelmember.HasUserWith(user.ID(botID))).Exec(ctx); err != nil {
		return err
	}
	return client.App.DeleteOneID(appID).Exec(ctx)
}

func permissionStrings(permissions []entity.AppPermission) []string {
	result := make([]string, len(permissions))
	for i, p := range permissions {
		result[i] = string(p)
	}
	return result
}

func appToEntity(a *ent.App) *entity.App {
	result := &entity.App{
		ID:             a.ID.String(),
		WorkspaceID:    a.WorkspaceID,
		Name:           a.Name,
		Description:    a.Description,
		AvatarURL:      a.AvatarURL,
		TokenHash:      a.TokenHash,
		OutgoingURL:    a.OutgoingURL,
		OutgoingSecret: a.OutgoingSecret,
		IsOfficial:     a.IsOfficial,
		LastUsedAt:     a.LastUsedAt,
		CreatedAt:      a.CreatedAt,
	}
	for _, p := range a.Permissions {
		result.Permissions = append(result.Permissions, entity.AppPermission(p))
	}
	if a.Edges.CreatedBy != nil {
		result.CreatedBy = a.Edges.CreatedBy.ID.String()
	}
	if a.Edges.BotUser != nil {
		result.BotUserID = a.Edges.BotUser.ID.String()
	}
	if a.Edges.DefaultChannel != nil {
		id := a.Edges.DefaultChannel.ID.String()
		result.DefaultChannelID = &id
	}
	return result
}
