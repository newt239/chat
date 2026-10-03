package service

import (
	"context"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type ChannelAccessService interface {
	EnsureChannelAccess(ctx context.Context, channelID string, userID string) (*entity.Channel, error)
	// EnsureMessageAccess はメッセージがあり、そのチャンネルを閲覧できることを確かめます
	EnsureMessageAccess(ctx context.Context, messageID string, userID string) (*entity.Message, *entity.Channel, error)
	// EnsureChannelMember は閲覧権限に加えて、チャンネルに参加していることを確かめます
	EnsureChannelMember(ctx context.Context, channelID string, userID string) (*entity.Channel, error)
	// FilterAccessible は公開チャンネルと参加中の非公開チャンネルだけを残します
	FilterAccessible(ctx context.Context, channels []*entity.Channel, userID string) ([]*entity.Channel, error)
	// AccessibleDescendants は閲覧できる子孫チャンネルを返します
	AccessibleDescendants(ctx context.Context, ch *entity.Channel, userID string) ([]*entity.Channel, error)
	// AccessibleChannelsByIDs は channelIDs のうち userID が閲覧できるチャンネルを ID ごとに返します
	AccessibleChannelsByIDs(ctx context.Context, channelIDs []string, userID string) (map[string]*entity.Channel, error)
	// FilterUsersWithAccess は userIDs のうちチャンネルを閲覧できるユーザーを返します
	FilterUsersWithAccess(ctx context.Context, ch *entity.Channel, userIDs []string) (map[string]bool, error)
}

type channelAccessService struct {
	messageRepo       domainrepository.MessageRepository
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	workspaceRepo     domainrepository.WorkspaceRepository
}

func NewChannelAccessService(
	messageRepo domainrepository.MessageRepository,
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
) ChannelAccessService {
	return &channelAccessService{
		messageRepo:       messageRepo,
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		workspaceRepo:     workspaceRepo,
	}
}

func (s *channelAccessService) EnsureChannelAccess(ctx context.Context, channelID string, userID string) (*entity.Channel, error) {
	ch, err := s.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to load channel: %w", err)
	}
	if ch == nil {
		return nil, domerr.ErrChannelNotFound
	}

	// 停止中のメンバーは除外されるため非公開チャンネルでも先に確認する
	if _, err := EnsureMember(ctx, s.workspaceRepo, ch.WorkspaceID, userID); err != nil {
		return nil, err
	}

	if ch.IsPrivate() {
		isMember, err := s.channelMemberRepo.IsMember(ctx, ch.ID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify channel membership: %w", err)
		}
		if !isMember {
			return nil, domerr.ErrUnauthorized
		}
	}

	return ch, nil
}

func (s *channelAccessService) EnsureMessageAccess(ctx context.Context, messageID string, userID string) (*entity.Message, *entity.Channel, error) {
	msg, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load message: %w", err)
	}
	if msg == nil {
		return nil, nil, domerr.ErrMessageNotFound
	}
	ch, err := s.EnsureChannelAccess(ctx, msg.ChannelID, userID)
	if err != nil {
		return nil, nil, err
	}
	return msg, ch, nil
}

func (s *channelAccessService) EnsureChannelMember(ctx context.Context, channelID string, userID string) (*entity.Channel, error) {
	ch, err := s.EnsureChannelAccess(ctx, channelID, userID)
	if err != nil {
		return nil, err
	}
	// 非公開チャンネルは EnsureChannelAccess で参加を確認済み
	if ch.IsPrivate() {
		return ch, nil
	}
	isMember, err := s.channelMemberRepo.IsMember(ctx, ch.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify channel membership: %w", err)
	}
	if !isMember {
		return nil, domerr.ErrNotChannelMember
	}
	return ch, nil
}

func (s *channelAccessService) FilterAccessible(ctx context.Context, channels []*entity.Channel, userID string) ([]*entity.Channel, error) {
	var privateIDs []string
	for _, ch := range channels {
		if ch.IsPrivate() {
			privateIDs = append(privateIDs, ch.ID)
		}
	}
	joined := map[string]bool{}
	if len(privateIDs) > 0 {
		var err error
		if joined, err = s.channelMemberRepo.FindJoinedChannelIDs(ctx, userID, privateIDs); err != nil {
			return nil, fmt.Errorf("failed to verify channel membership: %w", err)
		}
	}
	result := make([]*entity.Channel, 0, len(channels))
	for _, ch := range channels {
		if !ch.IsPrivate() || joined[ch.ID] {
			result = append(result, ch)
		}
	}
	return result, nil
}

func (s *channelAccessService) AccessibleChannelsByIDs(ctx context.Context, channelIDs []string, userID string) (map[string]*entity.Channel, error) {
	result := map[string]*entity.Channel{}
	if len(channelIDs) == 0 {
		return result, nil
	}
	channels, err := s.channelRepo.FindByIDs(ctx, channelIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load channels: %w", err)
	}
	// 停止中のメンバーは FindMember で除外される。チャンネルは同じワークスペースにあることが多いため 1 回ずつ確かめる
	memberOf := map[string]bool{}
	var inWorkspace []*entity.Channel
	for _, ch := range channels {
		isMember, checked := memberOf[ch.WorkspaceID]
		if !checked {
			member, err := s.workspaceRepo.FindMember(ctx, ch.WorkspaceID, userID)
			if err != nil {
				return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
			}
			isMember = member != nil
			memberOf[ch.WorkspaceID] = isMember
		}
		if isMember {
			inWorkspace = append(inWorkspace, ch)
		}
	}
	accessible, err := s.FilterAccessible(ctx, inWorkspace, userID)
	if err != nil {
		return nil, err
	}
	for _, ch := range accessible {
		result[ch.ID] = ch
	}
	return result, nil
}

func (s *channelAccessService) FilterUsersWithAccess(ctx context.Context, ch *entity.Channel, userIDs []string) (map[string]bool, error) {
	if len(userIDs) == 0 {
		return map[string]bool{}, nil
	}
	members, err := s.workspaceRepo.FindActiveMemberIDs(ctx, ch.WorkspaceID, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if !ch.IsPrivate() {
		return members, nil
	}
	joined, err := s.channelMemberRepo.FindMemberIDsIn(ctx, ch.ID, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to verify channel membership: %w", err)
	}
	for id := range members {
		if !joined[id] {
			delete(members, id)
		}
	}
	return members, nil
}

func (s *channelAccessService) AccessibleDescendants(ctx context.Context, ch *entity.Channel, userID string) ([]*entity.Channel, error) {
	descendants, err := s.channelRepo.FindDescendants(ctx, ch)
	if err != nil {
		return nil, fmt.Errorf("failed to load descendant channels: %w", err)
	}
	return s.FilterAccessible(ctx, descendants, userID)
}
