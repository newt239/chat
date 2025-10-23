package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/example/chat/internal/domain"
	"github.com/example/chat/internal/infrastructure/db"
)

type channelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) domain.ChannelRepository {
	return &channelRepository{db: db}
}

func (r *channelRepository) FindByID(id string) (*domain.Channel, error) {
	channelID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid channel ID format")
	}

	var dbChannel db.Channel
	if err := r.db.Where("id = ?", channelID).First(&dbChannel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toChannelDomain(&dbChannel), nil
}

func (r *channelRepository) FindByWorkspaceID(workspaceID string) ([]*domain.Channel, error) {
	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, errors.New("invalid workspace ID format")
	}

	var dbChannels []db.Channel
	if err := r.db.Where("workspace_id = ?", wsID).Order("created_at asc").Find(&dbChannels).Error; err != nil {
		return nil, err
	}

	channels := make([]*domain.Channel, len(dbChannels))
	for i, c := range dbChannels {
		channels[i] = toChannelDomain(&c)
	}

	return channels, nil
}

func (r *channelRepository) FindAccessibleChannels(workspaceID, userID string) ([]*domain.Channel, error) {
	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, errors.New("invalid workspace ID format")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Find all public channels + private channels where user is a member
	var dbChannels []db.Channel

	// Subquery for private channels where user is a member
	subQuery := r.db.
		Table("channel_members").
		Select("channel_id").
		Where("user_id = ?", uid)

	if err := r.db.
		Where("workspace_id = ? AND (is_private = false OR id IN (?))", wsID, subQuery).
		Order("created_at asc").
		Find(&dbChannels).Error; err != nil {
		return nil, err
	}

	channels := make([]*domain.Channel, len(dbChannels))
	for i, c := range dbChannels {
		channels[i] = toChannelDomain(&c)
	}

	return channels, nil
}

func (r *channelRepository) Create(channel *domain.Channel) error {
	workspaceID, err := uuid.Parse(channel.WorkspaceID)
	if err != nil {
		return errors.New("invalid workspace ID format")
	}

	createdBy, err := uuid.Parse(channel.CreatedBy)
	if err != nil {
		return errors.New("invalid created_by user ID format")
	}

	dbChannel := &db.Channel{
		WorkspaceID: workspaceID,
		Name:        channel.Name,
		Description: channel.Description,
		IsPrivate:   channel.IsPrivate,
		CreatedBy:   createdBy,
	}

	if channel.ID != "" {
		channelID, err := uuid.Parse(channel.ID)
		if err != nil {
			return errors.New("invalid channel ID format")
		}
		dbChannel.ID = channelID
	}

	if err := r.db.Create(dbChannel).Error; err != nil {
		return err
	}

	channel.ID = dbChannel.ID.String()
	channel.CreatedAt = dbChannel.CreatedAt
	channel.UpdatedAt = dbChannel.UpdatedAt

	return nil
}

func (r *channelRepository) Update(channel *domain.Channel) error {
	channelID, err := uuid.Parse(channel.ID)
	if err != nil {
		return errors.New("invalid channel ID format")
	}

	updates := map[string]interface{}{
		"name":        channel.Name,
		"description": channel.Description,
		"is_private":  channel.IsPrivate,
	}

	if err := r.db.Model(&db.Channel{}).Where("id = ?", channelID).Updates(updates).Error; err != nil {
		return err
	}

	// Fetch updated record
	var updated db.Channel
	if err := r.db.Where("id = ?", channelID).First(&updated).Error; err != nil {
		return err
	}

	channel.UpdatedAt = updated.UpdatedAt

	return nil
}

func (r *channelRepository) Delete(id string) error {
	channelID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid channel ID format")
	}

	return r.db.Delete(&db.Channel{}, "id = ?", channelID).Error
}

func (r *channelRepository) AddMember(member *domain.ChannelMember) error {
	channelID, err := uuid.Parse(member.ChannelID)
	if err != nil {
		return errors.New("invalid channel ID format")
	}

	userID, err := uuid.Parse(member.UserID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	dbMember := &db.ChannelMember{
		ChannelID: channelID,
		UserID:    userID,
	}

	return r.db.Create(dbMember).Error
}

func (r *channelRepository) RemoveMember(channelID, userID string) error {
	chID, err := uuid.Parse(channelID)
	if err != nil {
		return errors.New("invalid channel ID format")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	return r.db.Delete(&db.ChannelMember{}, "channel_id = ? AND user_id = ?", chID, uid).Error
}

func (r *channelRepository) FindMembers(channelID string) ([]*domain.ChannelMember, error) {
	chID, err := uuid.Parse(channelID)
	if err != nil {
		return nil, errors.New("invalid channel ID format")
	}

	var dbMembers []db.ChannelMember
	if err := r.db.Where("channel_id = ?", chID).Order("joined_at asc").Find(&dbMembers).Error; err != nil {
		return nil, err
	}

	members := make([]*domain.ChannelMember, len(dbMembers))
	for i, m := range dbMembers {
		members[i] = toChannelMemberDomain(&m)
	}

	return members, nil
}

func (r *channelRepository) IsMember(channelID, userID string) (bool, error) {
	chID, err := uuid.Parse(channelID)
	if err != nil {
		return false, errors.New("invalid channel ID format")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, errors.New("invalid user ID format")
	}

	var count int64
	if err := r.db.Model(&db.ChannelMember{}).
		Where("channel_id = ? AND user_id = ?", chID, uid).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func toChannelDomain(dbChannel *db.Channel) *domain.Channel {
	return &domain.Channel{
		ID:          dbChannel.ID.String(),
		WorkspaceID: dbChannel.WorkspaceID.String(),
		Name:        dbChannel.Name,
		Description: dbChannel.Description,
		IsPrivate:   dbChannel.IsPrivate,
		CreatedBy:   dbChannel.CreatedBy.String(),
		CreatedAt:   dbChannel.CreatedAt,
		UpdatedAt:   dbChannel.UpdatedAt,
	}
}

func toChannelMemberDomain(dbMember *db.ChannelMember) *domain.ChannelMember {
	return &domain.ChannelMember{
		ChannelID: dbMember.ChannelID.String(),
		UserID:    dbMember.UserID.String(),
		JoinedAt:  dbMember.JoinedAt,
	}
}
