package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	channellinkuc "github.com/newt239/chat/internal/usecase/channellink"
)

type ChannelLinkServer struct {
	UC *channellinkuc.Interactor
}

func (s *ChannelLinkServer) ListChannelLinks(ctx context.Context, req *chatv1.ListChannelLinksRequest) (*chatv1.ListChannelLinksResponse, error) {
	links, canEdit, err := s.UC.List(ctx, req.ChannelId, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListChannelLinksResponse{Links: presenter.ConvertAll(links, presenter.ChannelLink), CanEdit: canEdit}, nil
}

func (s *ChannelLinkServer) CreateChannelLink(ctx context.Context, req *chatv1.CreateChannelLinkRequest) (*chatv1.CreateChannelLinkResponse, error) {
	out, err := s.UC.Create(ctx, channellinkuc.LinkInput{ID: req.ChannelId, UserID: userIDFrom(ctx), Title: req.Title, URL: req.Url})
	if err != nil {
		return nil, err
	}
	return &chatv1.CreateChannelLinkResponse{Link: presenter.ChannelLink(out)}, nil
}

func (s *ChannelLinkServer) UpdateChannelLink(ctx context.Context, req *chatv1.UpdateChannelLinkRequest) (*chatv1.UpdateChannelLinkResponse, error) {
	out, err := s.UC.Update(ctx, channellinkuc.LinkInput{ID: req.LinkId, UserID: userIDFrom(ctx), Title: req.Title, URL: req.Url})
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdateChannelLinkResponse{Link: presenter.ChannelLink(out)}, nil
}

func (s *ChannelLinkServer) DeleteChannelLink(ctx context.Context, req *chatv1.DeleteChannelLinkRequest) (*chatv1.DeleteChannelLinkResponse, error) {
	return &chatv1.DeleteChannelLinkResponse{}, s.UC.Delete(ctx, req.LinkId, userIDFrom(ctx))
}

func (s *ChannelLinkServer) ReorderChannelLinks(ctx context.Context, req *chatv1.ReorderChannelLinksRequest) (*chatv1.ReorderChannelLinksResponse, error) {
	return &chatv1.ReorderChannelLinksResponse{}, s.UC.Reorder(ctx, req.ChannelId, userIDFrom(ctx), req.LinkIds)
}
