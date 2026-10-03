package repository

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/predicate"
	"github.com/newt239/chat/ent/user"
	"github.com/newt239/chat/ent/userlink"
	"github.com/newt239/chat/ent/userpreference"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type userRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) domainrepository.UserRepository {
	return &userRepository{client: client}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	userID, err := parseUUID(id, "user ID")
	if err != nil {
		return nil, err
	}
	return r.findOne(ctx, user.ID(userID))
}

func (r *userRepository) FindByIDs(ctx context.Context, ids []string) (map[string]*entity.User, error) {
	userIDs, err := parseUUIDs(slices.Compact(slices.Sorted(slices.Values(ids))), "user ID")
	if err != nil {
		return nil, err
	}
	users, err := withProfile(transaction.ResolveClient(ctx, r.client).User.Query().Where(user.IDIn(userIDs...))).All(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*entity.User, len(users))
	for _, u := range users {
		result[u.ID.String()] = userToEntity(u)
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
	u, err := orNil(withProfile(transaction.ResolveClient(ctx, r.client).User.Query().Where(where)).Only(ctx))
	if u == nil {
		return nil, err
	}
	return userToEntity(u), nil
}

// withProfile は設定とプロフィールのリンクを読み込みます
func withProfile(q *ent.UserQuery) *ent.UserQuery {
	return q.WithPreference().WithLinks(func(q *ent.UserLinkQuery) { q.Order(ent.Asc(userlink.FieldPosition)) })
}

func (r *userRepository) Create(ctx context.Context, usr *entity.User) error {
	u, err := transaction.ResolveClient(ctx, r.client).User.Create().
		SetNillableID(parseUUIDPtr(&usr.ID)).
		SetEmail(usr.Email).
		SetPasswordHash(usr.PasswordHash).
		SetNillableGoogleSub(usr.GoogleSub).
		SetDisplayName(usr.DisplayName).
		SetNillableAvatarURL(usr.AvatarURL).
		SetIsApp(usr.IsApp).
		SetIsOfficial(usr.IsOfficial).
		Save(ctx)
	if err != nil {
		return err
	}
	*usr = *userToEntity(u)
	return nil
}

// Update はユーザーと設定・リンクを 1 つのトランザクションで書き換えます。設定は初めて書くときに作ります
func (r *userRepository) Update(ctx context.Context, usr *entity.User) error {
	userID, err := parseUUID(usr.ID, "user ID")
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
			builder.SetAvatarURL(*usr.AvatarURL)
		} else {
			builder.ClearAvatarURL()
		}
		if err := builder.Exec(ctx); err != nil {
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
		links := make([]*ent.UserLinkCreate, len(usr.Links))
		for i, url := range usr.Links {
			links[i] = client.UserLink.Create().SetUserID(userID).SetPosition(i).SetURL(url)
		}
		if err := client.UserLink.CreateBulk(links...).Exec(ctx); err != nil {
			return err
		}

		return nil
	})
}

// personalDataColumns は退会時に消す本人だけのデータの表と、ユーザーを指す列です
var personalDataColumns = [][2]string{
	{"session", "user_id"}, {"push_token", "user_id"}, {"workspace_member", "user_id"}, {"channel_member", "user_id"},
	{"channel_read_state", "user_id"}, {"thread_read_state", "user_id"}, {"user_thread_follow", "user_id"},
	{"channel_star", "user_id"}, {"channel_mute", "user_id"}, {"channel_category_item", "user_id"}, {"channel_category", "user_id"},
	{"draft", "user_id"}, {"scheduled_message", "user_id"}, {"reminder", "creator_id"}, {"reminder", "target_user_id"},
	{"message_bookmark", "user_id"}, {"user_note", "owner_id"}, {"user_group_member", "user_id"},
	{"user_link", "user_id"}, {"user_preference", "user_id"},
}

// Delete は本人だけのデータを消し、投稿などの名義として残す行を匿名化して退会済みにします
func (r *userRepository) Delete(ctx context.Context, id string) error {
	userID, err := parseUUID(id, "user ID")
	if err != nil {
		return err
	}
	return transaction.WithTx(ctx, r.client, func(client *ent.Client) error {
		for _, tc := range personalDataColumns {
			if _, err := client.ExecContext(ctx, fmt.Sprintf(`DELETE FROM %q WHERE %q = $1`, tc[0], tc[1]), userID); err != nil {
				return err
			}
		}
		return client.User.UpdateOneID(userID).
			SetEmail("deleted-" + id + "@deleted.invalid").
			SetPasswordHash(entity.UnusablePasswordHash).
			ClearGoogleSub().
			SetDisplayName(entity.DeletedUserDisplayName).
			ClearBio().
			ClearAvatarURL().
			SetDeletedAt(time.Now()).
			Exec(ctx)
	})
}
