package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/example/chat/internal/domain"
	"github.com/example/chat/internal/infrastructure/db"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(id string) (*domain.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	var dbUser db.User
	if err := r.db.Where("id = ?", userID).First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toUserDomain(&dbUser), nil
}

func (r *userRepository) FindByIDs(ids []string) ([]*domain.User, error) {
	if len(ids) == 0 {
		return []*domain.User{}, nil
	}

	userIDs := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		userID, err := uuid.Parse(id)
		if err != nil {
			return nil, errors.New("invalid user ID format")
		}
		userIDs = append(userIDs, userID)
	}

	var dbUsers []db.User
	if err := r.db.Where("id IN ?", userIDs).Find(&dbUsers).Error; err != nil {
		return nil, err
	}

	users := make([]*domain.User, 0, len(dbUsers))
	for i := range dbUsers {
		users = append(users, toUserDomain(&dbUsers[i]))
	}

	return users, nil
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var dbUser db.User
	if err := r.db.Where("email = ?", email).First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toUserDomain(&dbUser), nil
}

func (r *userRepository) Create(user *domain.User) error {
	dbUser := &db.User{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		DisplayName:  user.DisplayName,
		AvatarURL:    user.AvatarURL,
	}

	if user.ID != "" {
		userID, err := uuid.Parse(user.ID)
		if err != nil {
			return errors.New("invalid user ID format")
		}
		dbUser.ID = userID
	}

	if err := r.db.Create(dbUser).Error; err != nil {
		return err
	}

	// Update the domain object with generated values
	user.ID = dbUser.ID.String()
	user.CreatedAt = dbUser.CreatedAt
	user.UpdatedAt = dbUser.UpdatedAt

	return nil
}

func (r *userRepository) Update(user *domain.User) error {
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	dbUser := &db.User{
		ID:           userID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		DisplayName:  user.DisplayName,
		AvatarURL:    user.AvatarURL,
	}

	if err := r.db.Model(&db.User{}).Where("id = ?", userID).Updates(dbUser).Error; err != nil {
		return err
	}

	// Fetch updated record to get UpdatedAt
	var updated db.User
	if err := r.db.Where("id = ?", userID).First(&updated).Error; err != nil {
		return err
	}

	user.UpdatedAt = updated.UpdatedAt

	return nil
}

func (r *userRepository) Delete(id string) error {
	userID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	return r.db.Delete(&db.User{}, "id = ?", userID).Error
}

func toUserDomain(dbUser *db.User) *domain.User {
	return &domain.User{
		ID:           dbUser.ID.String(),
		Email:        dbUser.Email,
		PasswordHash: dbUser.PasswordHash,
		DisplayName:  dbUser.DisplayName,
		AvatarURL:    dbUser.AvatarURL,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}
}
