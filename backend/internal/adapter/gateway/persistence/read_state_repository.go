package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/example/chat/internal/domain/entity"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"github.com/example/chat/internal/infrastructure/database"
)

type readStateRepository struct {
	db *gorm.DB
}

func NewReadStateRepository(db *gorm.DB) domainrepository.ReadStateRepository {
	return &readStateRepository{db: db}
}

func (r *readStateRepository) FindByChannelAndUser(ctx context.Context, channelID string, userID string) (*entity.ChannelReadState, error) {
	chID, err := parseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}

	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	var model database.ChannelReadState
	if err := r.db.WithContext(ctx).Where("channel_id = ? AND user_id = ?", chID, uid).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToEntity(), nil
}

func (r *readStateRepository) Upsert(ctx context.Context, readState *entity.ChannelReadState) error {
	channelID, err := parseUUID(readState.ChannelID, "channel ID")
	if err != nil {
		return err
	}

	userID, err := parseUUID(readState.UserID, "user ID")
	if err != nil {
		return err
	}

	model := &database.ChannelReadState{}
	model.FromEntity(readState)
	model.ChannelID = channelID
	model.UserID = userID

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "channel_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_read_at"}),
	}).Create(model).Error
}

func (r *readStateRepository) GetUnreadCount(ctx context.Context, channelID string, userID string) (int, error) {
	chID, err := parseUUID(channelID, "channel ID")
	if err != nil {
		return 0, err
	}

	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return 0, err
	}

	var readState database.ChannelReadState
	err = r.db.WithContext(ctx).Where("channel_id = ? AND user_id = ?", chID, uid).First(&readState).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var count int64
			if err := r.db.WithContext(ctx).Model(&database.Message{}).
				Where("channel_id = ? AND deleted_at IS NULL", chID).
				Count(&count).Error; err != nil {
				return 0, err
			}
			return int(count), nil
		}
		return 0, err
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&database.Message{}).
		Where("channel_id = ? AND created_at > ? AND deleted_at IS NULL", chID, readState.LastReadAt).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}

func (r *readStateRepository) GetUnreadChannels(ctx context.Context, userID string) (map[string]int, error) {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	var channelIDs []uuid.UUID

	var workspaceIDs []uuid.UUID
	if err := r.db.WithContext(ctx).Table("workspace_members").
		Select("workspace_id").
		Where("user_id = ?", uid).
		Pluck("workspace_id", &workspaceIDs).Error; err != nil {
		return nil, err
	}

	publicChannelQuery := r.db.WithContext(ctx).Table("channels").
		Select("id").
		Where("workspace_id IN ? AND is_private = false", workspaceIDs)

	privateChannelQuery := r.db.WithContext(ctx).Table("channel_members").
		Select("channel_id").
		Where("user_id = ?", uid)

	if err := r.db.WithContext(ctx).Raw("(?) UNION (?)", publicChannelQuery, privateChannelQuery).
		Scan(&channelIDs).Error; err != nil {
		return nil, err
	}

	result := make(map[string]int)

	for _, chID := range channelIDs {
		count, err := r.GetUnreadCount(ctx, chID.String(), userID)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			result[chID.String()] = count
		}
	}

	return result, nil
}
