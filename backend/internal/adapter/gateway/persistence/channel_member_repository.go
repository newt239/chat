package persistence

import (
	"context"

	"gorm.io/gorm"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/database"
)

type channelMemberRepository struct {
	db *gorm.DB
}

func NewChannelMemberRepository(db *gorm.DB) domainrepository.ChannelMemberRepository {
	return &channelMemberRepository{db: db}
}

func (r *channelMemberRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return resolveDB(ctx, r.db)
}

func (r *channelMemberRepository) AddMember(ctx context.Context, member *entity.ChannelMember) error {
	model := &database.ChannelMember{}
	model.FromEntity(member)

	return r.dbWithContext(ctx).Create(model).Error
}

func (r *channelMemberRepository) RemoveMember(ctx context.Context, channelID string, userID string) error {
	return r.dbWithContext(ctx).Delete(&database.ChannelMember{}, "channel_id = ? AND user_id = ?", channelID, userID).Error
}

func (r *channelMemberRepository) FindMembers(ctx context.Context, channelID string) ([]*entity.ChannelMember, error) {
	var models []database.ChannelMember
	if err := r.dbWithContext(ctx).Where("channel_id = ?", channelID).Order("joined_at asc").Find(&models).Error; err != nil {
		return nil, err
	}

	members := make([]*entity.ChannelMember, len(models))
	for i, model := range models {
		members[i] = model.ToEntity()
	}

	return members, nil
}

func (r *channelMemberRepository) IsMember(ctx context.Context, channelID string, userID string) (bool, error) {
	var count int64
	if err := r.dbWithContext(ctx).Model(&database.ChannelMember{}).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *channelMemberRepository) UpdateMemberRole(ctx context.Context, channelID string, userID string, role entity.ChannelRole) error {
	return r.dbWithContext(ctx).Model(&database.ChannelMember{}).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Update("role", string(role)).Error
}

func (r *channelMemberRepository) CountAdmins(ctx context.Context, channelID string) (int, error) {
	var count int64
	if err := r.dbWithContext(ctx).Model(&database.ChannelMember{}).
		Where("channel_id = ? AND role = ?", channelID, string(entity.ChannelRoleAdmin)).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}
