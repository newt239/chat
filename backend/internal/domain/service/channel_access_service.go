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
	result := make([]*entity.Channel, 0, len(channels))
	for _, ch := range channels {
		if ch.IsPrivate() {
			isMember, err := s.channelMemberRepo.IsMember(ctx, ch.ID, userID)
			if err != nil {
				return nil, fmt.Errorf("failed to verify channel membership: %w", err)
			}
			if !isMember {
				continue
			}
		}
		result = append(result, ch)
	}
	return result, nil
}

func (s *channelAccessService) AccessibleDescendants(ctx context.Context, ch *entity.Channel, userID string) ([]*entity.Channel, error) {
	descendants, err := s.channelRepo.FindDescendants(ctx, ch)
	if err != nil {
		return nil, fmt.Errorf("failed to load descendant channels: %w", err)
	}
	return s.FilterAccessible(ctx, descendants, userID)
}
