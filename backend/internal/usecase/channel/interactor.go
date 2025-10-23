package channel

import (
	"errors"
	"fmt"
	"time"

	"github.com/example/chat/internal/domain"
)

var (
	ErrUnauthorized      = errors.New("unauthorized to perform this action")
	ErrWorkspaceNotFound = errors.New("workspace not found")
)

type ChannelUseCase interface {
	ListChannels(input ListChannelsInput) ([]ChannelOutput, error)
	CreateChannel(input CreateChannelInput) (*ChannelOutput, error)
}

type channelInteractor struct {
	channelRepo   domain.ChannelRepository
	workspaceRepo domain.WorkspaceRepository
}

func NewChannelInteractor(
	channelRepo domain.ChannelRepository,
	workspaceRepo domain.WorkspaceRepository,
) ChannelUseCase {
	return &channelInteractor{
		channelRepo:   channelRepo,
		workspaceRepo: workspaceRepo,
	}
}

func (i *channelInteractor) ListChannels(input ListChannelsInput) ([]ChannelOutput, error) {
	workspace, err := i.workspaceRepo.FindByID(input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace: %w", err)
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	member, err := i.workspaceRepo.FindMember(input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	channels, err := i.channelRepo.FindAccessibleChannels(input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channels: %w", err)
	}

	output := make([]ChannelOutput, 0, len(channels))
	for _, ch := range channels {
		output = append(output, toChannelOutput(ch))
	}

	return output, nil
}

func (i *channelInteractor) CreateChannel(input CreateChannelInput) (*ChannelOutput, error) {
	workspace, err := i.workspaceRepo.FindByID(input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace: %w", err)
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	member, err := i.workspaceRepo.FindMember(input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify membership: %w", err)
	}
	if member == nil || (member.Role != domain.WorkspaceRoleOwner && member.Role != domain.WorkspaceRoleAdmin) {
		return nil, ErrUnauthorized
	}

	channel := &domain.Channel{
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Description: input.Description,
		IsPrivate:   input.IsPrivate,
		CreatedBy:   input.UserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := i.channelRepo.Create(channel); err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	// Add creator as member if channel is private
	if channel.IsPrivate {
		member := &domain.ChannelMember{
			ChannelID: channel.ID,
			UserID:    input.UserID,
			JoinedAt:  time.Now(),
		}
		if err := i.channelRepo.AddMember(member); err != nil {
			return nil, fmt.Errorf("failed to add creator to private channel: %w", err)
		}
	}

	output := toChannelOutput(channel)
	return &output, nil
}

func toChannelOutput(channel *domain.Channel) ChannelOutput {
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
