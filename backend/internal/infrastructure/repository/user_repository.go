package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/user"
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
	u, err := client.User.Query().
		Where(user.ID(userID)).
		Only(ctx)
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
	users, err := client.User.Query().
		Where(user.IDIn(userIDs...)).
		All(ctx)
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
	client := transaction.ResolveClient(ctx, r.client)
	u, err := client.User.Query().
		Where(user.Email(email)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return utils.UserToEntity(u), nil
}

func (r *userRepository) Create(ctx context.Context, usr *entity.User) error {
	client := transaction.ResolveClient(ctx, r.client)

    builder := client.User.Create().
		SetEmail(usr.Email).
		SetPasswordHash(usr.PasswordHash).
		SetDisplayName(usr.DisplayName)

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

    if usr.Bio != nil {
        builder = builder.SetBio(*usr.Bio)
    }

	u, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	*usr = *utils.UserToEntity(u)
	return nil
}

func (r *userRepository) Update(ctx context.Context, usr *entity.User) error {
	userID, err := utils.ParseUUID(usr.ID, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

    builder := client.User.UpdateOneID(userID).
		SetEmail(usr.Email).
		SetPasswordHash(usr.PasswordHash).
		SetDisplayName(usr.DisplayName)

	if usr.AvatarURL != nil {
		builder = builder.SetAvatarURL(*usr.AvatarURL)
	} else {
		builder = builder.ClearAvatarURL()
	}

    if usr.Bio != nil {
        builder = builder.SetBio(*usr.Bio)
    } else {
        builder = builder.ClearBio()
    }

	u, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	usr.UpdatedAt = u.UpdatedAt
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	userID, err := utils.ParseUUID(id, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	return client.User.DeleteOneID(userID).Exec(ctx)
}
