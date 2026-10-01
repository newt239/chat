package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	channeluc "github.com/newt239/chat/internal/usecase/channel"
)

type ChannelServer struct {
	UC channeluc.ChannelUseCase
}

func (s *ChannelServer) ListChannels(ctx context.Context, req *chatv1.ListChannelsRequest) (*chatv1.ListChannelsResponse, error) {
	out, err := s.UC.ListChannels(ctx, channeluc.ListChannelsInput{WorkspaceID: req.WorkspaceId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.ListChannelsResponse{Channels: presenter.ConvertAll(out, presenter.Channel)}, nil
}

func (s *ChannelServer) ListBrowsableChannels(ctx context.Context, req *chatv1.ListBrowsableChannelsRequest) (*chatv1.ListBrowsableChannelsResponse, error) {
	out, err := s.UC.ListBrowsableChannels(ctx, channeluc.ListChannelsInput{WorkspaceID: req.WorkspaceId, UserID: userIDFrom(ctx)})
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
		Membership:  presenter.BrowsableChannelMembership(req.Membership),
		Sort:        presenter.BrowsableChannelSort(req.Sort),
		Page:        int(req.Page),
		PerPage:     int(req.PerPage),
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.SearchBrowsableChannelsResponse{
		Channels: presenter.ConvertAll(out.Channels, presenter.BrowsableChannel),
		Total:    int32(out.Total),
	}, nil
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
	out, err := s.UC.GetChannel(ctx, channeluc.GetChannelInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx)})
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

func (s *ChannelServer) ArchiveChannel(ctx context.Context, req *chatv1.ArchiveChannelRequest) (*chatv1.ArchiveChannelResponse, error) {
	out, err := s.UC.SetArchived(ctx, channeluc.SetArchivedInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx), Archived: true})
	if err != nil {
		return nil, err
	}
	return &chatv1.ArchiveChannelResponse{Channel: presenter.Channel(*out)}, nil
}

func (s *ChannelServer) UnarchiveChannel(ctx context.Context, req *chatv1.UnarchiveChannelRequest) (*chatv1.UnarchiveChannelResponse, error) {
	out, err := s.UC.SetArchived(ctx, channeluc.SetArchivedInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx), Archived: false})
	if err != nil {
		return nil, err
	}
	return &chatv1.UnarchiveChannelResponse{Channel: presenter.Channel(*out)}, nil
}

func (s *ChannelServer) DeleteChannel(ctx context.Context, req *chatv1.DeleteChannelRequest) (*chatv1.DeleteChannelResponse, error) {
	if err := s.UC.DeleteChannel(ctx, channeluc.DeleteChannelInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx)}); err != nil {
		return nil, err
	}
	return &chatv1.DeleteChannelResponse{}, nil
}

func (s *ChannelServer) SetChannelStarred(ctx context.Context, req *chatv1.SetChannelStarredRequest) (*chatv1.SetChannelStarredResponse, error) {
	err := s.UC.SetChannelStarred(ctx, channeluc.SetChannelStarredInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx), Starred: req.Starred})
	if err != nil {
		return nil, err
	}
	return &chatv1.SetChannelStarredResponse{}, nil
}

func (s *ChannelServer) SetChannelMuted(ctx context.Context, req *chatv1.SetChannelMutedRequest) (*chatv1.SetChannelMutedResponse, error) {
	err := s.UC.SetChannelMuted(ctx, channeluc.SetChannelMutedInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx), Muted: req.Muted})
	if err != nil {
		return nil, err
	}
	return &chatv1.SetChannelMutedResponse{}, nil
}
