package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/app"
	"github.com/newt239/chat/ent/channelmember"
	"github.com/newt239/chat/ent/user"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type appRepository struct {
	client *ent.Client
}

func NewAppRepository(client *ent.Client) domainrepository.AppRepository {
	return &appRepository{client: client}
}

func (r *appRepository) query(ctx context.Context) *ent.AppQuery {
	return transaction.ResolveClient(ctx, r.client).App.Query()
}

func (r *appRepository) find(ctx context.Context, query *ent.AppQuery) ([]*entity.App, error) {
	found, err := query.Order(ent.Asc(app.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(found, appToEntity), nil
}

func (r *appRepository) findOne(ctx context.Context, query *ent.AppQuery) (*entity.App, error) {
	found, err := orNil(query.Only(ctx))
	if found == nil {
		return nil, err
	}
	return appToEntity(found), nil
}

func (r *appRepository) FindByID(ctx context.Context, id string) (*entity.App, error) {
	appID, err := parseUUID(id, "app ID")
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
	cid, err := parseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}
	return r.find(ctx, r.query(ctx).Where(app.HasBotUserWith(
		user.HasChannelMembersWith(channelmember.ChannelID(cid)),
	)))
}

func (r *appRepository) Create(ctx context.Context, a *entity.App) error {
	creatorID, err := parseUUID(a.CreatedBy, "user ID")
	if err != nil {
		return err
	}
	botID, err := parseUUID(a.BotUserID, "bot user ID")
	if err != nil {
		return err
	}
	created, err := transaction.ResolveClient(ctx, r.client).App.Create().
		SetWorkspaceID(a.WorkspaceID).
		SetCreatedByID(creatorID).
		SetBotUserID(botID).
		SetIsOfficial(a.IsOfficial).
		SetName(a.Name).
		SetNillableDescription(a.Description).
		SetNillableAvatarURL(a.AvatarURL).
		SetNillableTokenHash(a.TokenHash).
		SetPermissions(permissionStrings(a.Permissions)).
		SetNillableDefaultChannelID(parseUUIDPtr(a.DefaultChannelID)).
		SetNillableOutgoingURL(a.OutgoingURL).
		SetNillableOutgoingSecret(a.OutgoingSecret).
		Save(ctx)
	if err != nil {
		return err
	}
	a.ID = created.ID.String()
	a.CreatedAt = created.CreatedAt
	return nil
}

func (r *appRepository) Update(ctx context.Context, a *entity.App) error {
	appID, err := parseUUID(a.ID, "app ID")
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
	if id := parseUUIDPtr(a.DefaultChannelID); id != nil {
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
	appID, err := parseUUID(id, "app ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).App.UpdateOneID(appID).SetLastUsedAt(usedAt).Exec(ctx)
}

// Delete はアプリと、ボットユーザーのチャンネルへの参加を消します。ボットユーザーは過去の投稿の名義として残す
func (r *appRepository) Delete(ctx context.Context, id string) error {
	appID, err := parseUUID(id, "app ID")
	if err != nil {
		return err
	}
	return transaction.WithTx(ctx, r.client, func(client *ent.Client) error {
		botID, err := client.App.Query().Where(app.ID(appID)).QueryBotUser().OnlyID(ctx)
		if err != nil {
			return err
		}
		if _, err := client.ChannelMember.Delete().Where(channelmember.UserID(botID)).Exec(ctx); err != nil {
			return err
		}
		return client.App.DeleteOneID(appID).Exec(ctx)
	})
}

func permissionStrings(permissions []entity.AppPermission) []string {
	return convertAll(permissions, func(p entity.AppPermission) string { return string(p) })
}

func appToEntity(a *ent.App) *entity.App {
	return &entity.App{
		ID:               a.ID.String(),
		WorkspaceID:      a.WorkspaceID,
		Name:             a.Name,
		Description:      a.Description,
		AvatarURL:        a.AvatarURL,
		TokenHash:        a.TokenHash,
		OutgoingURL:      a.OutgoingURL,
		OutgoingSecret:   a.OutgoingSecret,
		IsOfficial:       a.IsOfficial,
		LastUsedAt:       a.LastUsedAt,
		CreatedBy:        a.CreatedByID.String(),
		BotUserID:        a.BotUserID.String(),
		DefaultChannelID: optionalString(a.DefaultChannelID),
		Permissions:      convertAll(a.Permissions, func(p string) entity.AppPermission { return entity.AppPermission(p) }),
		CreatedAt:        a.CreatedAt,
	}
}
