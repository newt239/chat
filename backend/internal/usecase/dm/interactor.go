package dm

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/domain/repository"
)

type Interactor struct {
	channelRepo       repository.ChannelRepository
	channelMemberRepo repository.ChannelMemberRepository
	userRepo          repository.UserRepository
}

func NewInteractor(
	channelRepo repository.ChannelRepository,
	channelMemberRepo repository.ChannelMemberRepository,
	userRepo repository.UserRepository,
) *Interactor {
	return &Interactor{
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		userRepo:          userRepo,
	}
}

func (i *Interactor) CreateDM(ctx context.Context, input CreateDMInput) (*DMOutput, error) {
	targetUser, err := i.userRepo.FindByID(ctx, input.TargetUserID)
	if err != nil {
		return nil, err
	}
	if targetUser == nil {
		return nil, entity.ErrUserNotFound
	}

	channel, err := i.channelRepo.FindOrCreateDM(ctx, input.WorkspaceID, input.UserID, input.TargetUserID)
	if err != nil {
		return nil, err
	}

	members, err := i.channelMemberRepo.FindMembers(ctx, channel.ID)
	if err != nil {
		return nil, err
	}

	if len(members) == 0 {
		if err := i.channelMemberRepo.AddMember(ctx, &entity.ChannelMember{
			ChannelID: channel.ID,
			UserID:    input.UserID,
			Role:      entity.ChannelRoleMember,
			JoinedAt:  time.Now().UTC(),
		}); err != nil {
			return nil, err
		}

		if err := i.channelMemberRepo.AddMember(ctx, &entity.ChannelMember{
			ChannelID: channel.ID,
			UserID:    input.TargetUserID,
			Role:      entity.ChannelRoleMember,
			JoinedAt:  time.Now().UTC(),
		}); err != nil {
			return nil, err
		}
	}

	return i.buildDMOutput(ctx, channel, input.UserID)
}

func (i *Interactor) CreateGroupDM(ctx context.Context, input CreateGroupDMInput) (*DMOutput, error) {
	if len(input.MemberIDs) > 9 {
		return nil, entity.ErrGroupDMMaxMembers
	}

	users, err := i.userRepo.FindByIDs(ctx, input.MemberIDs)
	if err != nil {
		return nil, err
	}

	if len(users) != len(input.MemberIDs) {
		return nil, entity.ErrUserNotFound
	}

	channel, err := i.channelRepo.FindOrCreateGroupDM(ctx, input.WorkspaceID, input.CreatorID, input.MemberIDs, input.Name)
	if err != nil {
		return nil, err
	}

	existingMembers, err := i.channelMemberRepo.FindMembers(ctx, channel.ID)
	if err != nil {
		return nil, err
	}

	if len(existingMembers) == 0 {
		for _, memberID := range input.MemberIDs {
			if err := i.channelMemberRepo.AddMember(ctx, &entity.ChannelMember{
				ChannelID: channel.ID,
				UserID:    memberID,
				Role:      entity.ChannelRoleMember,
				JoinedAt:  time.Now().UTC(),
			}); err != nil {
				return nil, err
			}
		}
	}

	return i.buildDMOutput(ctx, channel, input.CreatorID)
}

func (i *Interactor) ListDMs(ctx context.Context, input ListDMsInput) ([]*DMOutput, error) {
	channels, err := i.channelRepo.FindUserDMs(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, err
	}

	result := make([]*DMOutput, 0, len(channels))
	for _, ch := range channels {
		output, err := i.buildDMOutput(ctx, ch, input.RequestUserID)
		if err != nil {
			return nil, err
		}
		result = append(result, output)
	}

	return result, nil
}

func (i *Interactor) buildDMOutput(ctx context.Context, channel *entity.Channel, requestUserID string) (*DMOutput, error) {
	members, err := i.channelMemberRepo.FindMembers(ctx, channel.ID)
	if err != nil {
		return nil, err
	}

	memberOutputs := make([]DMMemberOutput, 0, len(members))
	for _, member := range members {
		if member.UserID == requestUserID {
			continue
		}
		user, err := i.userRepo.FindByID(ctx, member.UserID)
		if err != nil {
			return nil, err
		}
		if user != nil {
			memberOutputs = append(memberOutputs, DMMemberOutput{
				UserID:      user.ID,
				DisplayName: user.DisplayName,
				AvatarURL:   user.AvatarURL,
			})
		}
	}

	return &DMOutput{
		ID:          channel.ID,
		WorkspaceID: channel.WorkspaceID,
		Name:        channel.Name,
		Description: channel.Description,
		Type:        string(channel.Type),
		Members:     memberOutputs,
		CreatedAt:   channel.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   channel.UpdatedAt.Format(time.RFC3339),
	}, nil
}
