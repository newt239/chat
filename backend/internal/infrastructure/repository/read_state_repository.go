package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/example/chat/internal/domain"
	"github.com/example/chat/internal/infrastructure/db"
)

type readStateRepository struct {
	db *gorm.DB
}

func NewReadStateRepository(db *gorm.DB) domain.ReadStateRepository {
	return &readStateRepository{db: db}
}

func (r *readStateRepository) FindByChannelAndUser(channelID, userID string) (*domain.ChannelReadState, error) {
	chID, err := uuid.Parse(channelID)
	if err != nil {
		return nil, errors.New("invalid channel ID format")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	var dbReadState db.ChannelReadState
	if err := r.db.Where("channel_id = ? AND user_id = ?", chID, uid).First(&dbReadState).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toReadStateDomain(&dbReadState), nil
}

func (r *readStateRepository) Upsert(readState *domain.ChannelReadState) error {
	channelID, err := uuid.Parse(readState.ChannelID)
	if err != nil {
		return errors.New("invalid channel ID format")
	}

	userID, err := uuid.Parse(readState.UserID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	dbReadState := &db.ChannelReadState{
		ChannelID:  channelID,
		UserID:     userID,
		LastReadAt: readState.LastReadAt,
	}

	// Use ON CONFLICT to upsert
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "channel_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_read_at"}),
	}).Create(dbReadState).Error
}

func (r *readStateRepository) GetUnreadCount(channelID, userID string) (int, error) {
	chID, err := uuid.Parse(channelID)
	if err != nil {
		return 0, errors.New("invalid channel ID format")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return 0, errors.New("invalid user ID format")
	}

	// Find the last read timestamp for this user/channel
	var dbReadState db.ChannelReadState
	err = r.db.Where("channel_id = ? AND user_id = ?", chID, uid).First(&dbReadState).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No read state exists - count all messages
			var count int64
			if err := r.db.Model(&db.Message{}).
				Where("channel_id = ? AND deleted_at IS NULL", chID).
				Count(&count).Error; err != nil {
				return 0, err
			}
			return int(count), nil
		}
		return 0, err
	}

	// Count messages after last read
	var count int64
	if err := r.db.Model(&db.Message{}).
		Where("channel_id = ? AND created_at > ? AND deleted_at IS NULL", chID, dbReadState.LastReadAt).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}

func (r *readStateRepository) GetUnreadChannels(userID string) (map[string]int, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Get all channels the user is a member of
	var channelIDs []uuid.UUID

	// Get channels from workspace membership (public channels) + channel membership (private channels)
	// First, get all workspaces user is a member of
	var workspaceIDs []uuid.UUID
	if err := r.db.Table("workspace_members").
		Select("workspace_id").
		Where("user_id = ?", uid).
		Pluck("workspace_id", &workspaceIDs).Error; err != nil {
		return nil, err
	}

	// Get all public channels in those workspaces
	publicChannelQuery := r.db.Table("channels").
		Select("id").
		Where("workspace_id IN ? AND is_private = false", workspaceIDs)

	// Get all private channels user is explicitly a member of
	privateChannelQuery := r.db.Table("channel_members").
		Select("channel_id").
		Where("user_id = ?", uid)

	// Union both queries
	if err := r.db.Raw("(?) UNION (?)", publicChannelQuery, privateChannelQuery).
		Scan(&channelIDs).Error; err != nil {
		return nil, err
	}

	result := make(map[string]int)

	// For each channel, calculate unread count
	for _, chID := range channelIDs {
		count, err := r.GetUnreadCount(chID.String(), userID)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			result[chID.String()] = count
		}
	}

	return result, nil
}

func toReadStateDomain(dbReadState *db.ChannelReadState) *domain.ChannelReadState {
	return &domain.ChannelReadState{
		ChannelID:  dbReadState.ChannelID.String(),
		UserID:     dbReadState.UserID.String(),
		LastReadAt: dbReadState.LastReadAt,
	}
}
