package channel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domaintransaction "github.com/newt239/chat/internal/domain/transaction"
	"github.com/newt239/chat/internal/usecase/systemmessage"
)

var (
	ErrUnauthorized      = errors.New("この操作を行う権限がありません")
	ErrWorkspaceNotFound = errors.New("ワークスペースが見つかりません")
)

type ChannelUseCase interface {
	ListChannels(ctx context.Context, input ListChannelsInput) ([]ChannelOutput, error)
	CreateChannel(ctx context.Context, input CreateChannelInput) (*ChannelOutput, error)
	UpdateChannel(ctx context.Context, input UpdateChannelInput) (*ChannelOutput, error)
}

type channelInteractor struct {
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	workspaceRepo     domainrepository.WorkspaceRepository
	readStateRepo     domainrepository.ReadStateRepository
	txManager         domaintransaction.Manager
	systemMessageUC   systemmessage.UseCase
}

func NewChannelInteractor(
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	readStateRepo domainrepository.ReadStateRepository,
	txManager domaintransaction.Manager,
	systemMessageUC systemmessage.UseCase,
) ChannelUseCase {
	return &channelInteractor{
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		workspaceRepo:     workspaceRepo,
		readStateRepo:     readStateRepo,
		txManager:         txManager,
		systemMessageUC:   systemMessageUC,
	}
}

func (i *channelInteractor) ListChannels(ctx context.Context, input ListChannelsInput) ([]ChannelOutput, error) {
	if err := validateUUID(input.WorkspaceID, "workspace ID"); err != nil {
		return nil, err
	}
	if err := validateUUID(input.UserID, "user ID"); err != nil {
		return nil, err
	}

	workspace, err := i.workspaceRepo.FindByID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace: %w", err)
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	member, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	channels, err := i.channelRepo.FindAccessibleChannels(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channels: %w", err)
	}

	// チャネルIDリストを作成
	channelIDs := make([]string, len(channels))
	for idx, ch := range channels {
		channelIDs[idx] = ch.ID
	}

	// バッチでメンション数を取得
	mentionCounts, err := i.readStateRepo.GetUnreadMentionCountBatch(ctx, channelIDs, input.UserID)
	if err != nil {
		// エラーの場合はログに記録し、空のマップとして扱う
		fmt.Printf("[WARN] Failed to get unread mention counts: userID=%s err=%v\n", input.UserID, err)
		mentionCounts = make(map[string]int)
	}

	output := make([]ChannelOutput, 0, len(channels))
	for _, ch := range channels {
		mentionCount := mentionCounts[ch.ID]
		hasMention := mentionCount > 0
		output = append(output, toChannelOutputWithUnread(ch, hasMention, mentionCount))
	}

	return output, nil
}

func (i *channelInteractor) CreateChannel(ctx context.Context, input CreateChannelInput) (*ChannelOutput, error) {
	if err := validateUUID(input.WorkspaceID, "workspace ID"); err != nil {
		return nil, err
	}
	if err := validateUUID(input.UserID, "user ID"); err != nil {
		return nil, err
	}

	workspace, err := i.workspaceRepo.FindByID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace: %w", err)
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	member, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify membership: %w", err)
	}
	if member == nil || !member.CanCreateChannel() {
		return nil, ErrUnauthorized
	}

	channel, err := entity.NewChannel(entity.ChannelParams{
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Description: input.Description,
		IsPrivate:   input.IsPrivate,
		CreatedBy:   input.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create channel entity: %w", err)
	}

	err = i.txManager.Do(ctx, func(txCtx context.Context) error {
		if err := i.channelRepo.Create(txCtx, channel); err != nil {
			return fmt.Errorf("failed to create channel: %w", err)
		}

		if channel.IsPrivate {
			member := &entity.ChannelMember{
				ChannelID: channel.ID,
				UserID:    input.UserID,
				Role:      entity.ChannelRoleAdmin,
				JoinedAt:  time.Now(),
			}
			if err := i.channelMemberRepo.AddMember(txCtx, member); err != nil {
				return fmt.Errorf("failed to add creator to private channel: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	output := toChannelOutputWithUnread(channel, false, 0) // 新規作成時はメンション数0
	return &output, nil
}

func (i *channelInteractor) UpdateChannel(ctx context.Context, input UpdateChannelInput) (*ChannelOutput, error) {
	if err := validateUUID(input.ChannelID, "channel ID"); err != nil {
		return nil, err
	}
	if err := validateUUID(input.UserID, "user ID"); err != nil {
		return nil, err
	}

	ch, err := i.channelRepo.FindByID(ctx, input.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channel: %w", err)
	}
	if ch == nil {
		return nil, errors.New("チャンネルが見つかりません")
	}

	// 権限: ワークスペースの管理権限（チャンネル編集権限として流用）
	wsMember, err := i.workspaceRepo.FindMember(ctx, ch.WorkspaceID, input.UserID)
	if err != nil || wsMember == nil || !wsMember.CanCreateChannel() {
		return nil, ErrUnauthorized
	}

	// 変更適用
	originalName := ch.Name
	originalDesc := ch.Description
	originalPrivate := ch.IsPrivate

	nameChanged := false
	descChanged := false
	privChanged := false

	if input.Name != nil {
		_ = ch.ChangeName(*input.Name)
		nameChanged = (originalName != ch.Name)
	}
	if input.Description != nil {
		ch.Description = input.Description
		ch.UpdatedAt = time.Now().UTC()
		// detect change
		old := ""
		if originalDesc != nil {
			old = *originalDesc
		}
		now := ""
		if ch.Description != nil {
			now = *ch.Description
		}
		descChanged = (old != now)
	}
	if input.IsPrivate != nil {
		ch.IsPrivate = *input.IsPrivate
		ch.UpdatedAt = time.Now().UTC()
		privChanged = (originalPrivate != ch.IsPrivate)
	}

	if err := i.channelRepo.Update(ctx, ch); err != nil {
		return nil, fmt.Errorf("failed to update channel: %w", err)
	}

	// 変更に応じてシステムメッセージ作成
	actorID := input.UserID
	if i.systemMessageUC != nil {
		if nameChanged {
			if _, err := i.systemMessageUC.Create(ctx, systemmessage.CreateInput{
				ChannelID: ch.ID,
				Kind:      entity.SystemMessageKindChannelNameChanged,
				Payload:   map[string]any{"from": originalName, "to": ch.Name},
				ActorID:   &actorID,
			}); err != nil {
				fmt.Printf("[WARN] Failed to create system message for channel name change: channelID=%s err=%v\n", ch.ID, err)
			}
		}
		if descChanged {
			from := ""
			if originalDesc != nil {
				from = *originalDesc
			}
			to := ""
			if ch.Description != nil {
				to = *ch.Description
			}
			if _, err := i.systemMessageUC.Create(ctx, systemmessage.CreateInput{
				ChannelID: ch.ID,
				Kind:      entity.SystemMessageKindChannelDescriptionChanged,
				Payload:   map[string]any{"from": from, "to": to},
				ActorID:   &actorID,
			}); err != nil {
				fmt.Printf("[WARN] Failed to create system message for channel description change: channelID=%s err=%v\n", ch.ID, err)
			}
		}
		if privChanged {
			from := "public"
			if originalPrivate {
				from = "private"
			}
			to := "public"
			if ch.IsPrivate {
				to = "private"
			}
			if _, err := i.systemMessageUC.Create(ctx, systemmessage.CreateInput{
				ChannelID: ch.ID,
				Kind:      entity.SystemMessageKindChannelPrivacyChanged,
				Payload:   map[string]any{"from": from, "to": to},
				ActorID:   &actorID,
			}); err != nil {
				fmt.Printf("[WARN] Failed to create system message for channel privacy change: channelID=%s err=%v\n", ch.ID, err)
			}
		}
	}

	out := toChannelOutput(ch)
	// 補足: メンションはfalse/0で返す（一覧APIの責務と分離）
	out.HasMention = false
	out.MentionCount = 0
	return &out, nil
}

func toChannelOutput(channel *entity.Channel) ChannelOutput {
	return toChannelOutputWithUnread(channel, false, 0)
}

func toChannelOutputWithUnread(channel *entity.Channel, hasMention bool, mentionCount int) ChannelOutput {
	return ChannelOutput{
		ID:           channel.ID,
		WorkspaceID:  channel.WorkspaceID,
		Name:         channel.Name,
		Description:  channel.Description,
		IsPrivate:    channel.IsPrivate,
		CreatedBy:    channel.CreatedBy,
		CreatedAt:    channel.CreatedAt,
		UpdatedAt:    channel.UpdatedAt,
		HasMention:   hasMention,
		MentionCount: mentionCount,
	}
}

func validateUUID(id string, label string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("%w: invalid %s format", domerr.ErrValidation, label)
	}
	return nil
}
