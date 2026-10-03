package dm

import (
	"context"
	"errors"
	"slices"

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

func New(
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

// ensureWorkspaceMembers は指定ユーザーが全員停止されずにワークスペースに参加していることを確認します
func (i *Interactor) ensureWorkspaceMembers(ctx context.Context, workspaceID string, userIDs ...string) error {
	members, err := i.workspaceRepo.FindActiveMemberIDs(ctx, workspaceID, userIDs)
	if err != nil {
		return err
	}
	for _, userID := range userIDs {
		if !members[userID] {
			return domerr.ErrUnauthorized
		}
	}
	return nil
}

func (i *Interactor) CreateDM(ctx context.Context, input CreateDMInput) (*DMOutput, error) {
	if err := i.ensureWorkspaceMembers(ctx, input.WorkspaceID, input.UserID, input.TargetUserID); err != nil {
		return nil, err
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

func (i *Interactor) buildDMOutput(ctx context.Context, channel *entity.Channel, requestUserID string) (*DMOutput, error) {
	outputs, err := i.buildDMOutputs(ctx, []*entity.Channel{channel}, requestUserID)
	if err != nil {
		return nil, err
	}
	return outputs[0], nil
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
		err := i.channelMemberRepo.AddMember(ctx, &entity.ChannelMember{ChannelID: channelID, UserID: userID, Role: entity.ChannelRoleMember})
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

	result, err := i.buildDMOutputs(ctx, channels, input.UserID)
	if err != nil {
		return nil, err
	}
	for _, output := range result {
		output.IsStarred = starred[output.ID]
		output.IsMuted = muted[output.ID]
		output.UnreadCount = unreadCounts[output.ID]
	}
	return result, nil
}

// buildDMOutputs は自分以外の参加者をまとめて読み込んで DM の出力を作ります
func (i *Interactor) buildDMOutputs(ctx context.Context, channels []*entity.Channel, requestUserID string) ([]*DMOutput, error) {
	channelIDs := make([]string, len(channels))
	for idx, ch := range channels {
		channelIDs[idx] = ch.ID
	}
	members, err := i.channelMemberRepo.FindMembersByChannelIDs(ctx, channelIDs)
	if err != nil {
		return nil, err
	}
	var userIDs []string
	for _, m := range members {
		if m.UserID != requestUserID && !slices.Contains(userIDs, m.UserID) {
			userIDs = append(userIDs, m.UserID)
		}
	}
	users, err := i.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	outputs := make([]*DMOutput, 0, len(channels))
	byID := make(map[string]*DMOutput, len(channels))
	for _, ch := range channels {
		output := &DMOutput{Channel: ch}
		outputs = append(outputs, output)
		byID[ch.ID] = output
	}
	for _, m := range members {
		if u := users[m.UserID]; u != nil && m.UserID != requestUserID {
			output := byID[m.ChannelID]
			output.Members = append(output.Members, u)
		}
	}
	return outputs, nil
}
