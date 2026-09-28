package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	channellinkuc "github.com/newt239/chat/internal/usecase/channellink"
)

type ChannelLinkServer struct {
	UC channellinkuc.UseCase
}

func (s *ChannelLinkServer) ListChannelLinks(ctx context.Context, req *chatv1.ListChannelLinksRequest) (*chatv1.ListChannelLinksResponse, error) {
	out, err := s.UC.List(ctx, channellinkuc.ListInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.ListChannelLinksResponse{Links: presenter.ConvertAll(out.Links, presenter.ChannelLink), CanEdit: out.CanEdit}, nil
}

func (s *ChannelLinkServer) CreateChannelLink(ctx context.Context, req *chatv1.CreateChannelLinkRequest) (*chatv1.CreateChannelLinkResponse, error) {
	out, err := s.UC.Create(ctx, channellinkuc.CreateInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx), Title: req.Title, URL: req.Url})
	if err != nil {
		return nil, err
	}
	return &chatv1.CreateChannelLinkResponse{Link: presenter.ChannelLink(*out)}, nil
}

func (s *ChannelLinkServer) UpdateChannelLink(ctx context.Context, req *chatv1.UpdateChannelLinkRequest) (*chatv1.UpdateChannelLinkResponse, error) {
	out, err := s.UC.Update(ctx, channellinkuc.UpdateInput{LinkID: req.LinkId, UserID: userIDFrom(ctx), Title: req.Title, URL: req.Url})
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdateChannelLinkResponse{Link: presenter.ChannelLink(*out)}, nil
}

func (s *ChannelLinkServer) DeleteChannelLink(ctx context.Context, req *chatv1.DeleteChannelLinkRequest) (*chatv1.DeleteChannelLinkResponse, error) {
	if err := s.UC.Delete(ctx, channellinkuc.DeleteInput{LinkID: req.LinkId, UserID: userIDFrom(ctx)}); err != nil {
		return nil, err
	}
	return &chatv1.DeleteChannelLinkResponse{}, nil
}

func (s *ChannelLinkServer) ReorderChannelLinks(ctx context.Context, req *chatv1.ReorderChannelLinksRequest) (*chatv1.ReorderChannelLinksResponse, error) {
	out, err := s.UC.Reorder(ctx, channellinkuc.ReorderInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx), LinkIDs: req.LinkIds})
	if err != nil {
		return nil, err
	}
	return &chatv1.ReorderChannelLinksResponse{Links: presenter.ConvertAll(out, presenter.ChannelLink)}, nil
}
