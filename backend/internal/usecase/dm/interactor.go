package dm

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	"github.com/newt239/chat/internal/domain/repository"
)

type Interactor struct {
	channelRepo       repository.ChannelRepository
	channelMemberRepo repository.ChannelMemberRepository
	channelStarRepo   repository.ChannelStarRepository
	channelMuteRepo   repository.ChannelMuteRepository
	readStateRepo     repository.ReadStateRepository
	userRepo          repository.UserRepository
	workspaceRepo     repository.WorkspaceRepository
}

func NewInteractor(
	channelRepo repository.ChannelRepository,
	channelMemberRepo repository.ChannelMemberRepository,
	channelStarRepo repository.ChannelStarRepository,
	channelMuteRepo repository.ChannelMuteRepository,
	readStateRepo repository.ReadStateRepository,
	userRepo repository.UserRepository,
	workspaceRepo repository.WorkspaceRepository,
) *Interactor {
	return &Interactor{
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		channelStarRepo:   channelStarRepo,
		channelMuteRepo:   channelMuteRepo,
		readStateRepo:     readStateRepo,
		userRepo:          userRepo,
		workspaceRepo:     workspaceRepo,
	}
}

// ensureWorkspaceMembers は指定ユーザーが全員ワークスペースのメンバーであることを確認します
func (i *Interactor) ensureWorkspaceMembers(ctx context.Context, workspaceID string, userIDs ...string) error {
	for _, userID := range userIDs {
		member, err := i.workspaceRepo.FindMember(ctx, workspaceID, userID)
		if err != nil {
			return err
		}
		if member == nil {
			return domerr.ErrUnauthorized
		}
	}
	return nil
}

func (i *Interactor) CreateDM(ctx context.Context, input CreateDMInput) (*DMOutput, error) {
	if err := i.ensureWorkspaceMembers(ctx, input.WorkspaceID, input.UserID, input.TargetUserID); err != nil {
		return nil, err
	}

	targetUser, err := i.userRepo.FindByID(ctx, input.TargetUserID)
	if err != nil {
		return nil, err
	}
	if targetUser == nil {
		return nil, domerr.ErrUserNotFound
	}

	channel, err := i.channelRepo.FindOrCreateDM(ctx, input.WorkspaceID, input.UserID, input.TargetUserID)
	if err != nil {
		return nil, err
	}

	if err := i.joinMembers(ctx, channel.ID, input.UserID, input.TargetUserID); err != nil {
		return nil, err
	}
	return i.buildDMOutput(ctx, channel, input.UserID)
}

func (i *Interactor) CreateGroupDM(ctx context.Context, input CreateGroupDMInput) (*DMOutput, error) {
	// 作成者を必ずメンバーに含める
	if !slices.Contains(input.MemberIDs, input.CreatorID) {
		input.MemberIDs = append([]string{input.CreatorID}, input.MemberIDs...)
	}

	if len(input.MemberIDs) > entity.MaxGroupDMMembers {
		return nil, entity.ErrGroupDMMaxMembers
	}

	if err := i.ensureWorkspaceMembers(ctx, input.WorkspaceID, input.MemberIDs...); err != nil {
		return nil, err
	}

	users, err := i.userRepo.FindByIDs(ctx, input.MemberIDs)
	if err != nil {
		return nil, err
	}

	if len(users) != len(input.MemberIDs) {
		return nil, domerr.ErrUserNotFound
	}

	channel, err := i.channelRepo.FindOrCreateGroupDM(ctx, input.WorkspaceID, input.CreatorID, input.MemberIDs, input.Name)
	if err != nil {
		return nil, err
	}

	if err := i.joinMembers(ctx, channel.ID, input.MemberIDs...); err != nil {
		return nil, err
	}
	return i.buildDMOutput(ctx, channel, input.CreatorID)
}

// joinMembers は DM の参加者を揃えます。同時に作られて既に参加していても成功させる
func (i *Interactor) joinMembers(ctx context.Context, channelID string, userIDs ...string) error {
	for _, userID := range userIDs {
		err := i.channelMemberRepo.AddMember(ctx, &entity.ChannelMember{ChannelID: channelID, UserID: userID, Role: entity.ChannelRoleMember, JoinedAt: time.Now().UTC()})
		if err != nil && !errors.Is(err, domerr.ErrAlreadyMember) {
			return err
		}
	}
	return nil
}

func (i *Interactor) ListDMs(ctx context.Context, input ListDMsInput) ([]*DMOutput, error) {
	channels, err := i.channelRepo.FindUserDMs(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, err
	}

	channelIDs := make([]string, len(channels))
	for idx, ch := range channels {
		channelIDs[idx] = ch.ID
	}
	starred, err := i.channelStarRepo.FindStarredChannelIDs(ctx, input.UserID, channelIDs)
	if err != nil {
		return nil, err
	}
	muted, err := i.channelMuteRepo.FindMutedChannelIDs(ctx, input.UserID, channelIDs)
	if err != nil {
		return nil, err
	}
	unreadCounts, err := i.readStateRepo.GetUnreadCountBatch(ctx, channelIDs, input.UserID)
	if err != nil {
		return nil, err
	}
	mentionCounts, err := i.readStateRepo.GetUnreadMentionCountBatch(ctx, channelIDs, input.UserID)
	if err != nil {
		return nil, err
	}

	result := make([]*DMOutput, 0, len(channels))
	for _, ch := range channels {
		output, err := i.buildDMOutput(ctx, ch, input.UserID)
		if err != nil {
			return nil, err
		}
		output.IsStarred = starred[ch.ID]
		output.IsMuted = muted[ch.ID]
		output.UnreadCount = unreadCounts[ch.ID]
		output.HasMention = mentionCounts[ch.ID] > 0
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
