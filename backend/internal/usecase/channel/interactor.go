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
)

var (
	ErrUnauthorized      = errors.New("この操作を行う権限がありません")
	ErrWorkspaceNotFound = errors.New("ワークスペースが見つかりません")
)

type ChannelUseCase interface {
	ListChannels(ctx context.Context, input ListChannelsInput) ([]ChannelOutput, error)
	CreateChannel(ctx context.Context, input CreateChannelInput) (*ChannelOutput, error)
}

type channelInteractor struct {
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	workspaceRepo     domainrepository.WorkspaceRepository
	readStateRepo     domainrepository.ReadStateRepository
	txManager         domaintransaction.Manager
}

func NewChannelInteractor(
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	readStateRepo domainrepository.ReadStateRepository,
	txManager domaintransaction.Manager,
) ChannelUseCase {
	return &channelInteractor{
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		workspaceRepo:     workspaceRepo,
		readStateRepo:     readStateRepo,
		txManager:         txManager,
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

	output := make([]ChannelOutput, 0, len(channels))
	for _, ch := range channels {
		// 未読数を取得
		unreadCount, err := i.readStateRepo.GetUnreadCount(ctx, ch.ID, input.UserID)
		if err != nil {
			// エラーの場合は0として扱う
			unreadCount = 0
		}

		// TODO: メンション検知の実装（現在は未読数が0より大きい場合にtrueとする）
		hasMention := unreadCount > 0

		output = append(output, toChannelOutputWithUnread(ch, unreadCount, hasMention))
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

	output := toChannelOutputWithUnread(channel, 0, false) // 新規作成時は未読数0
	return &output, nil
}

func toChannelOutput(channel *entity.Channel) ChannelOutput {
	return toChannelOutputWithUnread(channel, 0, false)
}

func toChannelOutputWithUnread(channel *entity.Channel, unreadCount int, hasMention bool) ChannelOutput {
	return ChannelOutput{
		ID:          channel.ID,
		WorkspaceID: channel.WorkspaceID,
		Name:        channel.Name,
		Description: channel.Description,
		IsPrivate:   channel.IsPrivate,
		CreatedBy:   channel.CreatedBy,
		CreatedAt:   channel.CreatedAt,
		UpdatedAt:   channel.UpdatedAt,
		UnreadCount: unreadCount,
		HasMention:  hasMention,
	}
}

func validateUUID(id string, label string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("%w: invalid %s format", domerr.ErrValidation, label)
	}
	return nil
}
