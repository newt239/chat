package channel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/example/chat/internal/domain/entity"
	domainrepository "github.com/example/chat/internal/domain/repository"
)

var (
	ErrUnauthorized      = errors.New("unauthorized to perform this action")
	ErrWorkspaceNotFound = errors.New("workspace not found")
)

type ChannelUseCase interface {
	ListChannels(ctx context.Context, input ListChannelsInput) ([]ChannelOutput, error)
	CreateChannel(ctx context.Context, input CreateChannelInput) (*ChannelOutput, error)
}

type channelInteractor struct {
	channelRepo   domainrepository.ChannelRepository
	workspaceRepo domainrepository.WorkspaceRepository
}

func NewChannelInteractor(
	channelRepo domainrepository.ChannelRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
) ChannelUseCase {
	return &channelInteractor{
		channelRepo:   channelRepo,
		workspaceRepo: workspaceRepo,
	}
}

func (i *channelInteractor) ListChannels(ctx context.Context, input ListChannelsInput) ([]ChannelOutput, error) {
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
		output = append(output, toChannelOutput(ch))
	}

	return output, nil
}

func (i *channelInteractor) CreateChannel(ctx context.Context, input CreateChannelInput) (*ChannelOutput, error) {
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
	if member == nil || (member.Role != entity.WorkspaceRoleOwner && member.Role != entity.WorkspaceRoleAdmin) {
		return nil, ErrUnauthorized
	}

	channel := &entity.Channel{
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Description: input.Description,
		IsPrivate:   input.IsPrivate,
		CreatedBy:   input.UserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := i.channelRepo.Create(ctx, channel); err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	// Add creator as member if channel is private
	if channel.IsPrivate {
		member := &entity.ChannelMember{
			ChannelID: channel.ID,
			UserID:    input.UserID,
			JoinedAt:  time.Now(),
		}
		if err := i.channelRepo.AddMember(ctx, member); err != nil {
			return nil, fmt.Errorf("failed to add creator to private channel: %w", err)
		}
	}

	output := toChannelOutput(channel)
	return &output, nil
}

func toChannelOutput(channel *entity.Channel) ChannelOutput {
	return ChannelOutput{
		ID:          channel.ID,
		WorkspaceID: channel.WorkspaceID,
		Name:        channel.Name,
		Description: channel.Description,
		IsPrivate:   channel.IsPrivate,
		CreatedBy:   channel.CreatedBy,
		CreatedAt:   channel.CreatedAt,
		UpdatedAt:   channel.UpdatedAt,
	}
}
