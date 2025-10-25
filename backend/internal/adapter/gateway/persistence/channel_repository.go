package persistence

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/example/chat/internal/domain/entity"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"github.com/example/chat/internal/infrastructure/database"
)

type channelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) domainrepository.ChannelRepository {
	return &channelRepository{db: db}
}

func (r *channelRepository) FindByID(ctx context.Context, id string) (*entity.Channel, error) {
	channelID, err := parseUUID(id, "channel ID")
	if err != nil {
		return nil, err
	}

	var model database.Channel
	if err := r.db.WithContext(ctx).Where("id = ?", channelID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToEntity(), nil
}

func (r *channelRepository) FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.Channel, error) {
	wsID, err := parseUUID(workspaceID, "workspace ID")
	if err != nil {
		return nil, err
	}

	var models []database.Channel
	if err := r.db.WithContext(ctx).Where("workspace_id = ?", wsID).Order("created_at asc").Find(&models).Error; err != nil {
		return nil, err
	}

	channels := make([]*entity.Channel, len(models))
	for i, model := range models {
		channels[i] = model.ToEntity()
	}

	return channels, nil
}

func (r *channelRepository) FindAccessibleChannels(ctx context.Context, workspaceID string, userID string) ([]*entity.Channel, error) {
	wsID, err := parseUUID(workspaceID, "workspace ID")
	if err != nil {
		return nil, err
	}

	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	var models []database.Channel

	subQuery := r.db.WithContext(ctx).
		Table("channel_members").
		Select("channel_id").
		Where("user_id = ?", uid)

	if err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND (is_private = false OR id IN (?))", wsID, subQuery).
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
	workspaceID, err := parseUUID(channel.WorkspaceID, "workspace ID")
	if err != nil {
		return err
	}

	createdBy, err := parseUUID(channel.CreatedBy, "created_by user ID")
	if err != nil {
		return err
	}

	model := &database.Channel{}
	model.FromEntity(channel)
	model.WorkspaceID = workspaceID
	model.CreatedBy = createdBy

	if channel.ID != "" {
		channelID, err := parseUUID(channel.ID, "channel ID")
		if err != nil {
			return err
		}
		model.ID = channelID
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	*channel = *model.ToEntity()
	return nil
}

func (r *channelRepository) Update(ctx context.Context, channel *entity.Channel) error {
	channelID, err := parseUUID(channel.ID, "channel ID")
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"name":        channel.Name,
		"description": channel.Description,
		"is_private":  channel.IsPrivate,
	}

	if err := r.db.WithContext(ctx).Model(&database.Channel{}).Where("id = ?", channelID).Updates(updates).Error; err != nil {
		return err
	}

	var updated database.Channel
	if err := r.db.WithContext(ctx).Where("id = ?", channelID).First(&updated).Error; err != nil {
		return err
	}

	channel.UpdatedAt = updated.UpdatedAt
	return nil
}

func (r *channelRepository) Delete(ctx context.Context, id string) error {
	channelID, err := parseUUID(id, "channel ID")
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&database.Channel{}, "id = ?", channelID).Error
}

func (r *channelRepository) AddMember(ctx context.Context, member *entity.ChannelMember) error {
	channelID, err := parseUUID(member.ChannelID, "channel ID")
	if err != nil {
		return err
	}

	userID, err := parseUUID(member.UserID, "user ID")
	if err != nil {
		return err
	}

	model := &database.ChannelMember{}
	model.FromEntity(member)
	model.ChannelID = channelID
	model.UserID = userID

	return r.db.WithContext(ctx).Create(model).Error
}

func (r *channelRepository) RemoveMember(ctx context.Context, channelID string, userID string) error {
	chID, err := parseUUID(channelID, "channel ID")
	if err != nil {
		return err
	}

	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&database.ChannelMember{}, "channel_id = ? AND user_id = ?", chID, uid).Error
}

func (r *channelRepository) FindMembers(ctx context.Context, channelID string) ([]*entity.ChannelMember, error) {
	chID, err := parseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}

	var models []database.ChannelMember
	if err := r.db.WithContext(ctx).Where("channel_id = ?", chID).Order("joined_at asc").Find(&models).Error; err != nil {
		return nil, err
	}

	members := make([]*entity.ChannelMember, len(models))
	for i, model := range models {
		members[i] = model.ToEntity()
	}

	return members, nil
}

func (r *channelRepository) IsMember(ctx context.Context, channelID string, userID string) (bool, error) {
	chID, err := parseUUID(channelID, "channel ID")
	if err != nil {
		return false, err
	}

	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return false, err
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&database.ChannelMember{}).
		Where("channel_id = ? AND user_id = ?", chID, uid).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *channelRepository) UpdateMemberRole(ctx context.Context, channelID string, userID string, role entity.ChannelRole) error {
	chID, err := parseUUID(channelID, "channel ID")
	if err != nil {
		return err
	}

	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Model(&database.ChannelMember{}).
		Where("channel_id = ? AND user_id = ?", chID, uid).
		Update("role", string(role)).Error
}

func (r *channelRepository) CountAdmins(ctx context.Context, channelID string) (int, error) {
	chID, err := parseUUID(channelID, "channel ID")
	if err != nil {
		return 0, err
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&database.ChannelMember{}).
		Where("channel_id = ? AND role = ?", chID, string(entity.ChannelRoleAdmin)).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}
