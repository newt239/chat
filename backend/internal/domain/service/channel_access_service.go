package service

import (
    "context"
    "fmt"

    domainerrors "github.com/newt239/chat/internal/domain/errors"
    "github.com/newt239/chat/internal/domain/entity"
    domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type ChannelAccessService interface {
    EnsureChannelAccess(ctx context.Context, channelID string, userID string) (*entity.Channel, error)
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
        return nil, domainerrors.ErrChannelNotFound
    }

    if ch.IsPrivate {
        isMember, err := s.channelMemberRepo.IsMember(ctx, ch.ID, userID)
        if err != nil {
            return nil, fmt.Errorf("failed to verify channel membership: %w", err)
        }
        if !isMember {
            return nil, domainerrors.ErrUnauthorized
        }
        return ch, nil
    }

    member, err := s.workspaceRepo.FindMember(ctx, ch.WorkspaceID, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
    }
    if member == nil {
        return nil, domainerrors.ErrUnauthorized
    }

    return ch, nil
}


