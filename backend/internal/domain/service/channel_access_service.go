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
	// EnsureChannelMember は閲覧権限に加えて、チャンネルに参加していることを確かめます
	EnsureChannelMember(ctx context.Context, channelID string, userID string) (*entity.Channel, error)
	// FilterAccessible は公開チャンネルと参加中の非公開チャンネルだけを残します
	FilterAccessible(ctx context.Context, channels []*entity.Channel, userID string) ([]*entity.Channel, error)
	// AccessibleDescendants は閲覧できる子孫チャンネルを返します
	AccessibleDescendants(ctx context.Context, ch *entity.Channel, userID string) ([]*entity.Channel, error)
	// FilterUsersWithAccess は userIDs のうちチャンネルを閲覧できるユーザーを返します
	FilterUsersWithAccess(ctx context.Context, ch *entity.Channel, userIDs []string) (map[string]bool, error)
}

type channelAccessService struct {
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	workspaceRepo     domainrepository.WorkspaceRepository
}

func NewChannelAccessService(
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
) ChannelAccessService {
	return &channelAccessService{
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

	// 停止中のメンバーは FindMember で除外されるため非公開チャンネルでも先に確認する
	member, err := s.workspaceRepo.FindMember(ctx, ch.WorkspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, domerr.ErrUnauthorized
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

func (s *channelAccessService) EnsureChannelMember(ctx context.Context, channelID string, userID string) (*entity.Channel, error) {
	ch, err := s.EnsureChannelAccess(ctx, channelID, userID)
	if err != nil {
		return nil, err
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
