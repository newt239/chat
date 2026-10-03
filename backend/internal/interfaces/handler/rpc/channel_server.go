package rpc

import (
	"context"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	channeluc "github.com/newt239/chat/internal/usecase/channel"
)

type ChannelServer struct {
	UC *channeluc.Interactor
}

var browsableMemberships = map[chatv1.BrowsableChannelMembership]domainrepository.BrowsableChannelMembership{
	chatv1.BrowsableChannelMembership_BROWSABLE_CHANNEL_MEMBERSHIP_JOINED:     domainrepository.BrowsableChannelMembershipJoined,
	chatv1.BrowsableChannelMembership_BROWSABLE_CHANNEL_MEMBERSHIP_NOT_JOINED: domainrepository.BrowsableChannelMembershipNotJoined,
}

var browsableSorts = map[chatv1.BrowsableChannelSort]domainrepository.BrowsableChannelSort{
	chatv1.BrowsableChannelSort_BROWSABLE_CHANNEL_SORT_MEMBER_COUNT: domainrepository.BrowsableChannelSortMemberCount,
}

func (s *ChannelServer) ListChannels(ctx context.Context, req *chatv1.ListChannelsRequest) (*chatv1.ListChannelsResponse, error) {
	out, err := s.UC.ListChannels(ctx, req.WorkspaceId, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListChannelsResponse{Channels: presenter.ConvertAll(out, presenter.Channel)}, nil
}

func (s *ChannelServer) ListBrowsableChannels(ctx context.Context, req *chatv1.ListBrowsableChannelsRequest) (*chatv1.ListBrowsableChannelsResponse, error) {
	out, err := s.UC.ListBrowsableChannels(ctx, req.WorkspaceId, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListBrowsableChannelsResponse{Channels: presenter.ConvertAll(out, presenter.BrowsableChannel)}, nil
}

func (s *ChannelServer) SearchBrowsableChannels(ctx context.Context, req *chatv1.SearchBrowsableChannelsRequest) (*chatv1.SearchBrowsableChannelsResponse, error) {
	out, err := s.UC.SearchBrowsableChannels(ctx, channeluc.SearchBrowsableChannelsInput{
		WorkspaceID: req.WorkspaceId,
		UserID:      userIDFrom(ctx),
		Query:       req.Query,
		Membership:  browsableMemberships[req.Membership],
		Sort:        browsableSorts[req.Sort],
		Page:        int(req.Page),
		PerPage:     int(req.PerPage),
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.SearchBrowsableChannelsResponse{Channels: presenter.ConvertAll(out.Channels, presenter.BrowsableChannel), Total: int32(out.Total)}, nil
}

func (s *ChannelServer) CreateChannel(ctx context.Context, req *chatv1.CreateChannelRequest) (*chatv1.CreateChannelResponse, error) {
	out, err := s.UC.CreateChannel(ctx, channeluc.CreateChannelInput{
		WorkspaceID: req.WorkspaceId,
		UserID:      userIDFrom(ctx),
		Name:        req.Name,
		Description: req.Description,
		IsPrivate:   req.IsPrivate,
		MemberIDs:   req.MemberIds,
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.CreateChannelResponse{Channel: presenter.Channel(*out)}, nil
}

func (s *ChannelServer) GetChannel(ctx context.Context, req *chatv1.GetChannelRequest) (*chatv1.GetChannelResponse, error) {
	out, err := s.UC.GetChannel(ctx, req.ChannelId, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.GetChannelResponse{Channel: presenter.Channel(*out)}, nil
}

func (s *ChannelServer) UpdateChannel(ctx context.Context, req *chatv1.UpdateChannelRequest) (*chatv1.UpdateChannelResponse, error) {
	out, err := s.UC.UpdateChannel(ctx, channeluc.UpdateChannelInput{
		ChannelID:   req.ChannelId,
		UserID:      userIDFrom(ctx),
		Name:        req.Name,
		Description: req.Description,
		IsPrivate:   req.IsPrivate,
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdateChannelResponse{Channel: presenter.Channel(*out)}, nil
}

func (s *ChannelServer) SetChannelStarred(ctx context.Context, req *chatv1.SetChannelStarredRequest) (*chatv1.SetChannelStarredResponse, error) {
	return &chatv1.SetChannelStarredResponse{}, s.UC.SetChannelStarred(ctx, channeluc.SetFlagInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx), Value: req.Starred})
}

func (s *ChannelServer) SetChannelMuted(ctx context.Context, req *chatv1.SetChannelMutedRequest) (*chatv1.SetChannelMutedResponse, error) {
	return &chatv1.SetChannelMutedResponse{}, s.UC.SetChannelMuted(ctx, channeluc.SetFlagInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx), Value: req.Muted})
}
