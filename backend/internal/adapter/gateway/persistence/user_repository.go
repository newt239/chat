package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/example/chat/internal/domain/entity"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"github.com/example/chat/internal/infrastructure/database"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domainrepository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	userID, err := parseUUID(id, "user ID")
	if err != nil {
		return nil, err
	}

	var model database.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToEntity(), nil
}

func (r *userRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.User, error) {
	if len(ids) == 0 {
		return []*entity.User{}, nil
	}

	userIDs := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		userID, err := parseUUID(id, "user ID")
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	var models []database.User
	if err := r.db.WithContext(ctx).Where("id IN ?", userIDs).Find(&models).Error; err != nil {
		return nil, err
	}

	users := make([]*entity.User, 0, len(models))
	for i := range models {
		users = append(users, models[i].ToEntity())
	}

	return users, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var model database.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToEntity(), nil
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	model := &database.User{}
	model.FromEntity(user)

	if user.ID != "" {
		userID, err := parseUUID(user.ID, "user ID")
		if err != nil {
			return err
		}
		model.ID = userID
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	*user = *model.ToEntity()
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	userID, err := parseUUID(user.ID, "user ID")
	if err != nil {
		return err
	}

	model := &database.User{}
	model.FromEntity(user)
	model.ID = userID

	if err := r.db.WithContext(ctx).Model(&database.User{}).Where("id = ?", userID).Updates(model).Error; err != nil {
		return err
	}

	var updated database.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&updated).Error; err != nil {
		return err
	}

	user.UpdatedAt = updated.UpdatedAt
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	userID, err := parseUUID(id, "user ID")
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&database.User{}, "id = ?", userID).Error
}
