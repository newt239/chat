package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/predicate"
	"github.com/newt239/chat/ent/user"
	"github.com/newt239/chat/ent/userlink"
	"github.com/newt239/chat/ent/userpreference"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type userRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) domainrepository.UserRepository {
	return &userRepository{client: client}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	userID, err := utils.ParseUUID(id, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	u, err := withProfile(client.User.Query().Where(user.ID(userID))).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return utils.UserToEntity(u), nil
}

func (r *userRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.User, error) {
	if len(ids) == 0 {
		return []*entity.User{}, nil
	}

	userIDs := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		userID, err := utils.ParseUUID(id, "user ID")
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	client := transaction.ResolveClient(ctx, r.client)
	users, err := withProfile(client.User.Query().Where(user.IDIn(userIDs...))).All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.User, 0, len(users))
	for _, u := range users {
		result = append(result, utils.UserToEntity(u))
	}

	return result, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	return r.findOne(ctx, user.Email(email))
}

func (r *userRepository) FindByGoogleSub(ctx context.Context, sub string) (*entity.User, error) {
	return r.findOne(ctx, user.GoogleSub(sub))
}

func (r *userRepository) findOne(ctx context.Context, where predicate.User) (*entity.User, error) {
	client := transaction.ResolveClient(ctx, r.client)
	u, err := withProfile(client.User.Query().Where(where)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return utils.UserToEntity(u), nil
}

// withProfile は設定とプロフィールのリンクを読み込みます
func withProfile(q *ent.UserQuery) *ent.UserQuery {
	return q.WithPreference().WithLinks(func(q *ent.UserLinkQuery) { q.Order(ent.Asc(userlink.FieldPosition)) })
}

func (r *userRepository) Create(ctx context.Context, usr *entity.User) error {
	client := transaction.ResolveClient(ctx, r.client)

	builder := client.User.Create().
		SetEmail(usr.Email).
		SetPasswordHash(usr.PasswordHash).
		SetNillableGoogleSub(usr.GoogleSub).
		SetDisplayName(usr.DisplayName).
		SetIsApp(usr.IsApp).
		SetIsOfficial(usr.IsOfficial)

	if usr.ID != "" {
		userID, err := utils.ParseUUID(usr.ID, "user ID")
		if err != nil {
			return err
		}
		builder = builder.SetID(userID)
	}

	if usr.AvatarURL != nil {
		builder = builder.SetAvatarURL(*usr.AvatarURL)
	}

	u, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	*usr = *utils.UserToEntity(u)
	return nil
}

// Update はユーザーと設定・リンクを 1 つのトランザクションで書き換えます。設定は初めて書くときに作ります
func (r *userRepository) Update(ctx context.Context, usr *entity.User) error {
	userID, err := utils.ParseUUID(usr.ID, "user ID")
	if err != nil {
		return err
	}

	return transaction.WithTx(ctx, r.client, func(client *ent.Client) error {
		builder := client.User.UpdateOneID(userID).
			SetEmail(usr.Email).
			SetPasswordHash(usr.PasswordHash).
			SetNillableGoogleSub(usr.GoogleSub).
			SetDisplayName(usr.DisplayName).
			SetNillableBio(usr.Bio)
		if usr.AvatarURL != nil {
			builder = builder.SetAvatarURL(*usr.AvatarURL)
		} else {
			builder = builder.ClearAvatarURL()
		}
		u, err := builder.Save(ctx)
		if err != nil {
			return err
		}

		p := usr.Preferences
		if err := client.UserPreference.Create().
			SetUserID(userID).
			SetThemeHue(p.ThemeHue).
			SetThemeChroma(p.ThemeChroma).
			SetThemeSidebar(userpreference.ThemeSidebar(p.ThemeSidebar)).
			SetColorMode(userpreference.ColorMode(p.ColorMode)).
			SetLocale(p.Locale).
			SetNotificationLevel(userpreference.NotificationLevel(p.NotificationLevel)).
			SetTimezone(p.Timezone).
			SetTimezoneAutoUpdate(p.TimezoneAutoUpdate).
			SetChannelSortOrder(userpreference.ChannelSortOrder(p.ChannelSortOrder)).
			SetHideJoinMessages(p.HideJoinMessages).
			OnConflictColumns(userpreference.FieldUserID).
			UpdateNewValues().
			Exec(ctx); err != nil {
			return err
		}

		if _, err := client.UserLink.Delete().Where(userlink.UserID(userID)).Exec(ctx); err != nil {
			return err
		}
		links := make([]*ent.UserLinkCreate, 0, len(usr.Links))
		for i, url := range usr.Links {
			links = append(links, client.UserLink.Create().SetUserID(userID).SetPosition(i).SetURL(url))
		}
		if err := client.UserLink.CreateBulk(links...).Exec(ctx); err != nil {
			return err
		}

		usr.UpdatedAt = u.UpdatedAt
		return nil
	})
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	userID, err := utils.ParseUUID(id, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	return client.User.DeleteOneID(userID).Exec(ctx)
}
