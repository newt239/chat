package reaction

import (
	"errors"
	"fmt"
	"time"

	"github.com/example/chat/internal/domain"
)

var (
	ErrMessageNotFound = errors.New("message not found")
	ErrUnauthorized    = errors.New("unauthorized to perform this action")
	ErrReactionExists  = errors.New("reaction already exists")
)

type ReactionUseCase interface {
	AddReaction(input AddReactionInput) error
	RemoveReaction(input RemoveReactionInput) error
	ListReactions(messageID string, userID string) (*ListReactionsOutput, error)
}

type reactionInteractor struct {
	messageRepo   domain.MessageRepository
	channelRepo   domain.ChannelRepository
	workspaceRepo domain.WorkspaceRepository
	userRepo      domain.UserRepository
}

func NewReactionInteractor(
	messageRepo domain.MessageRepository,
	channelRepo domain.ChannelRepository,
	workspaceRepo domain.WorkspaceRepository,
	userRepo domain.UserRepository,
) ReactionUseCase {
	return &reactionInteractor{
		messageRepo:   messageRepo,
		channelRepo:   channelRepo,
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
	}
}

func (i *reactionInteractor) AddReaction(input AddReactionInput) error {
	// メッセージの存在確認とアクセス権限チェック
	message, err := i.messageRepo.FindByID(input.MessageID)
	if err != nil {
		return fmt.Errorf("failed to fetch message: %w", err)
	}
	if message == nil {
		return ErrMessageNotFound
	}

	// チャンネルへのアクセス権限チェック
	if err := i.ensureChannelAccess(message.ChannelID, input.UserID); err != nil {
		return err
	}

	// リアクションを追加
	reaction := &domain.MessageReaction{
		MessageID: input.MessageID,
		UserID:    input.UserID,
		Emoji:     input.Emoji,
		CreatedAt: time.Now(),
	}

	if err := i.messageRepo.AddReaction(reaction); err != nil {
		return fmt.Errorf("failed to add reaction: %w", err)
	}

	return nil
}

func (i *reactionInteractor) RemoveReaction(input RemoveReactionInput) error {
	// メッセージの存在確認
	message, err := i.messageRepo.FindByID(input.MessageID)
	if err != nil {
		return fmt.Errorf("failed to fetch message: %w", err)
	}
	if message == nil {
		return ErrMessageNotFound
	}

	// チャンネルへのアクセス権限チェック
	if err := i.ensureChannelAccess(message.ChannelID, input.UserID); err != nil {
		return err
	}

	// リアクションを削除
	if err := i.messageRepo.RemoveReaction(input.MessageID, input.UserID, input.Emoji); err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}

	return nil
}

func (i *reactionInteractor) ListReactions(messageID string, userID string) (*ListReactionsOutput, error) {
	// メッセージの存在確認
	message, err := i.messageRepo.FindByID(messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch message: %w", err)
	}
	if message == nil {
		return nil, ErrMessageNotFound
	}

	// チャンネルへのアクセス権限チェック
	if err := i.ensureChannelAccess(message.ChannelID, userID); err != nil {
		return nil, err
	}

	// リアクションを取得
	reactions, err := i.messageRepo.FindReactions(messageID)
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
	users, err := i.userRepo.FindByIDs(userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	// ユーザー情報をマップに格納
	userMap := make(map[string]*domain.User)
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

func (i *reactionInteractor) ensureChannelAccess(channelID, userID string) error {
	ch, err := i.channelRepo.FindByID(channelID)
	if err != nil {
		return fmt.Errorf("failed to load channel: %w", err)
	}
	if ch == nil {
		return errors.New("channel not found")
	}

	// プライベートチャンネルの場合
	if ch.IsPrivate {
		isMember, err := i.channelRepo.IsMember(ch.ID, userID)
		if err != nil {
			return fmt.Errorf("failed to verify channel membership: %w", err)
		}
		if !isMember {
			return ErrUnauthorized
		}
		return nil
	}

	// パブリックチャンネルの場合はワークスペースメンバーかチェック
	member, err := i.workspaceRepo.FindMember(ch.WorkspaceID, userID)
	if err != nil {
		return fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return ErrUnauthorized
	}

	return nil
}

func toReactionOutput(reaction *domain.MessageReaction, user *domain.User) ReactionOutput {
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
