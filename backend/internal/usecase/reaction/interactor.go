package reaction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/example/chat/internal/domain/entity"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"github.com/example/chat/internal/domain/service"
)

var (
	ErrMessageNotFound = errors.New("message not found")
	ErrUnauthorized    = errors.New("unauthorized to perform this action")
	ErrReactionExists  = errors.New("reaction already exists")
)

type ReactionUseCase interface {
	AddReaction(ctx context.Context, input AddReactionInput) error
	RemoveReaction(ctx context.Context, input RemoveReactionInput) error
	ListReactions(ctx context.Context, messageID string, userID string) (*ListReactionsOutput, error)
}

type reactionInteractor struct {
	messageRepo     domainrepository.MessageRepository
	channelRepo     domainrepository.ChannelRepository
	workspaceRepo   domainrepository.WorkspaceRepository
	userRepo        domainrepository.UserRepository
	notificationSvc service.NotificationService
}

func NewReactionInteractor(
	messageRepo domainrepository.MessageRepository,
	channelRepo domainrepository.ChannelRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
	notificationSvc service.NotificationService,
) ReactionUseCase {
	return &reactionInteractor{
		messageRepo:     messageRepo,
		channelRepo:     channelRepo,
		workspaceRepo:   workspaceRepo,
		userRepo:        userRepo,
		notificationSvc: notificationSvc,
	}
}

func (i *reactionInteractor) AddReaction(ctx context.Context, input AddReactionInput) error {
	// メッセージの存在確認とアクセス権限チェック
	message, err := i.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return fmt.Errorf("failed to fetch message: %w", err)
	}
	if message == nil {
		return ErrMessageNotFound
	}

	// チャンネルへのアクセス権限チェック
	if err := i.ensureChannelAccess(ctx, message.ChannelID, input.UserID); err != nil {
		return err
	}

	// リアクションを追加
	reaction := &entity.MessageReaction{
		MessageID: input.MessageID,
		UserID:    input.UserID,
		Emoji:     input.Emoji,
		CreatedAt: time.Now(),
	}

	if err := i.messageRepo.AddReaction(ctx, reaction); err != nil {
		return fmt.Errorf("failed to add reaction: %w", err)
	}

	// WebSocket通知を送信（nilチェックを追加）
	if i.notificationSvc != nil {
		// チャンネル情報を取得
		channel, err := i.channelRepo.FindByID(ctx, message.ChannelID)
		if err == nil && channel != nil {
			// reactionをmap[string]interface{}に変換
			reactionMap, err := convertStructToMap(toReactionOutput(reaction, nil))
			if err == nil {
				i.notificationSvc.NotifyReaction(channel.WorkspaceID, channel.ID, reactionMap)
			} else {
				fmt.Printf("Warning: failed to convert reaction to map: %v\n", err)
			}
		}
	}

	return nil
}

func (i *reactionInteractor) RemoveReaction(ctx context.Context, input RemoveReactionInput) error {
	// メッセージの存在確認
	message, err := i.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return fmt.Errorf("failed to fetch message: %w", err)
	}
	if message == nil {
		return ErrMessageNotFound
	}

	// チャンネルへのアクセス権限チェック
	if err := i.ensureChannelAccess(ctx, message.ChannelID, input.UserID); err != nil {
		return err
	}

	// リアクションを削除
	if err := i.messageRepo.RemoveReaction(ctx, input.MessageID, input.UserID, input.Emoji); err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}

	return nil
}

func (i *reactionInteractor) ListReactions(ctx context.Context, messageID string, userID string) (*ListReactionsOutput, error) {
	// メッセージの存在確認
	message, err := i.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch message: %w", err)
	}
	if message == nil {
		return nil, ErrMessageNotFound
	}

	// チャンネルへのアクセス権限チェック
	if err := i.ensureChannelAccess(ctx, message.ChannelID, userID); err != nil {
		return nil, err
	}

	// リアクションを取得
	reactions, err := i.messageRepo.FindReactions(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reactions: %w", err)
	}

	if len(reactions) == 0 {
		return &ListReactionsOutput{Reactions: []ReactionOutput{}}, nil
	}

	// ユーザーIDを収集
	userIDs := make([]string, 0, len(reactions))
	userIDSet := make(map[string]bool)
	for _, reaction := range reactions {
		if !userIDSet[reaction.UserID] {
			userIDs = append(userIDs, reaction.UserID)
			userIDSet[reaction.UserID] = true
		}
	}

	// ユーザー情報を一括取得
	users, err := i.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	// ユーザー情報をマップに格納
	userMap := make(map[string]*entity.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// ReactionOutputに変換
	outputs := make([]ReactionOutput, 0, len(reactions))
	for _, reaction := range reactions {
		user := userMap[reaction.UserID]
		outputs = append(outputs, toReactionOutput(reaction, user))
	}

	return &ListReactionsOutput{Reactions: outputs}, nil
}

func (i *reactionInteractor) ensureChannelAccess(ctx context.Context, channelID, userID string) error {
	ch, err := i.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to load channel: %w", err)
	}
	if ch == nil {
		return errors.New("channel not found")
	}

	// プライベートチャンネルの場合
	if ch.IsPrivate {
		isMember, err := i.channelRepo.IsMember(ctx, ch.ID, userID)
		if err != nil {
			return fmt.Errorf("failed to verify channel membership: %w", err)
		}
		if !isMember {
			return ErrUnauthorized
		}
		return nil
	}

	// パブリックチャンネルの場合はワークスペースメンバーかチェック
	member, err := i.workspaceRepo.FindMember(ctx, ch.WorkspaceID, userID)
	if err != nil {
		return fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return ErrUnauthorized
	}

	return nil
}

// convertStructToMap は構造体をmap[string]interface{}に変換します
func convertStructToMap(data interface{}) (map[string]interface{}, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func toReactionOutput(reaction *entity.MessageReaction, user *entity.User) ReactionOutput {
	userInfo := UserInfo{
		ID:          "",
		DisplayName: "Unknown User",
		AvatarURL:   nil,
	}

	if user != nil {
		userInfo = UserInfo{
			ID:          user.ID,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarURL,
		}
	}

	return ReactionOutput{
		MessageID: reaction.MessageID,
		User:      userInfo,
		Emoji:     reaction.Emoji,
		CreatedAt: reaction.CreatedAt,
	}
}
