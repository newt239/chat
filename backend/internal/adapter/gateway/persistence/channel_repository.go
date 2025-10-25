package persistence

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/database"
)

type channelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) domainrepository.ChannelRepository {
	return &channelRepository{db: db}
}

func (r *channelRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return resolveDB(ctx, r.db)
}

func (r *channelRepository) FindByID(ctx context.Context, id string) (*entity.Channel, error) {
	var model database.Channel
	if err := r.dbWithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToEntity(), nil
}

func (r *channelRepository) FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.Channel, error) {
	var models []database.Channel
	if err := r.dbWithContext(ctx).Where("workspace_id = ?", workspaceID).Order("created_at asc").Find(&models).Error; err != nil {
		return nil, err
	}

	channels := make([]*entity.Channel, len(models))
	for i, model := range models {
		channels[i] = model.ToEntity()
	}

	return channels, nil
}

func (r *channelRepository) FindAccessibleChannels(ctx context.Context, workspaceID string, userID string) ([]*entity.Channel, error) {
	var models []database.Channel

	subQuery := r.dbWithContext(ctx).
		Table("channel_members").
		Select("channel_id").
		Where("user_id = ?", userID)

	if err := r.dbWithContext(ctx).
		Where("workspace_id = ? AND (is_private = false OR id IN (?))", workspaceID, subQuery).
		Order("created_at asc").
		Find(&models).Error; err != nil {
		return nil, err
	}

	channels := make([]*entity.Channel, len(models))
	for i, model := range models {
		channels[i] = model.ToEntity()
	}

	return channels, nil
}

func (r *channelRepository) Create(ctx context.Context, channel *entity.Channel) error {
	model := &database.Channel{}
	model.FromEntity(channel)

	if err := r.dbWithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	*channel = *model.ToEntity()
	return nil
}

func (r *channelRepository) Update(ctx context.Context, channel *entity.Channel) error {
	type channelUpdate struct {
		Name        string
		Description *string
		IsPrivate   bool
	}

	if err := r.dbWithContext(ctx).
		Model(&database.Channel{}).
		Where("id = ?", channel.ID).
		Select("name", "description", "is_private").
		Updates(channelUpdate{
			Name:        channel.Name,
			Description: channel.Description,
			IsPrivate:   channel.IsPrivate,
		}).Error; err != nil {
		return err
	}

	var updated database.Channel
	if err := r.dbWithContext(ctx).Where("id = ?", channel.ID).First(&updated).Error; err != nil {
		return err
	}

	channel.UpdatedAt = updated.UpdatedAt
	return nil
}

func (r *channelRepository) Delete(ctx context.Context, id string) error {
	return r.dbWithContext(ctx).Delete(&database.Channel{}, "id = ?", id).Error
}
