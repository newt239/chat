package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/database"
)

type threadRepository struct {
	db *gorm.DB
}

func NewThreadRepository(db *gorm.DB) domainrepository.ThreadRepository {
	return &threadRepository{db: db}
}

func (r *threadRepository) FindMetadataByMessageID(ctx context.Context, messageID string) (*entity.ThreadMetadata, error) {
	msgID, err := parseUUID(messageID, "message ID")
	if err != nil {
		return nil, err
	}

	var model database.ThreadMetadata
	if err := r.db.WithContext(ctx).Where("message_id = ?", msgID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToEntity(), nil
}

func (r *threadRepository) FindMetadataByMessageIDs(ctx context.Context, messageIDs []string) (map[string]*entity.ThreadMetadata, error) {
	if len(messageIDs) == 0 {
		return make(map[string]*entity.ThreadMetadata), nil
	}

	msgIDs := make([]uuid.UUID, 0, len(messageIDs))
	for _, id := range messageIDs {
		msgID, err := parseUUID(id, "message ID")
		if err != nil {
			return nil, err
		}
		msgIDs = append(msgIDs, msgID)
	}

	var models []database.ThreadMetadata
	if err := r.db.WithContext(ctx).Where("message_id IN ?", msgIDs).Find(&models).Error; err != nil {
		return nil, err
	}

	result := make(map[string]*entity.ThreadMetadata, len(models))
	for _, model := range models {
		result[model.MessageID.String()] = model.ToEntity()
	}

	return result, nil
}

func (r *threadRepository) CreateOrUpdateMetadata(ctx context.Context, metadata *entity.ThreadMetadata) error {
	var model database.ThreadMetadata
	model.FromEntity(metadata)

	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "message_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"reply_count", "last_reply_at", "last_reply_user_id", "participant_user_ids", "updated_at"}),
	}).Create(&model).Error; err != nil {
		return err
	}

	return nil
}

func (r *threadRepository) IncrementReplyCount(ctx context.Context, messageID string, replyUserID string) error {
	msgID, err := parseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	replyUID, err := parseUUID(replyUserID, "reply user ID")
	if err != nil {
		return err
	}

	now := time.Now()

	// 既存のメタデータを取得
	var existing database.ThreadMetadata
	err = r.db.WithContext(ctx).Where("message_id = ?", msgID).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 新規作成
		newMetadata := database.ThreadMetadata{
			MessageID:          msgID,
			ReplyCount:         1,
			LastReplyAt:        &now,
			LastReplyUserID:    &replyUID,
			ParticipantUserIDs: []uuid.UUID{replyUID},
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		return r.db.WithContext(ctx).Create(&newMetadata).Error
	}

	// 参加者リストに追加（重複チェック）
	participants := existing.ParticipantUserIDs
	found := false
	for _, pid := range participants {
		if pid == replyUID {
			found = true
			break
		}
	}
	if !found {
		participants = append(participants, replyUID)
	}

	// 更新
	updates := map[string]interface{}{
		"reply_count":          gorm.Expr("reply_count + 1"),
		"last_reply_at":        now,
		"last_reply_user_id":   replyUID,
		"participant_user_ids": participants,
		"updated_at":           now,
	}

	return r.db.WithContext(ctx).Model(&database.ThreadMetadata{}).
		Where("message_id = ?", msgID).
		Updates(updates).Error
}

func (r *threadRepository) DeleteMetadata(ctx context.Context, messageID string) error {
	msgID, err := parseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&database.ThreadMetadata{}, "message_id = ?", msgID).Error
}
