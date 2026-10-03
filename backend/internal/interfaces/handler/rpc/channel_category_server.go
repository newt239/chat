package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	channelcategoryuc "github.com/newt239/chat/internal/usecase/channelcategory"
)

type ChannelCategoryServer struct {
	UC *channelcategoryuc.Interactor
}

func (s *ChannelCategoryServer) ListChannelCategories(ctx context.Context, req *chatv1.ListChannelCategoriesRequest) (*chatv1.ListChannelCategoriesResponse, error) {
	out, err := s.UC.List(ctx, req.WorkspaceId, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListChannelCategoriesResponse{Categories: presenter.ConvertAll(out, presenter.ChannelCategory)}, nil
}

func (s *ChannelCategoryServer) CreateChannelCategory(ctx context.Context, req *chatv1.CreateChannelCategoryRequest) (*chatv1.CreateChannelCategoryResponse, error) {
	out, err := s.UC.Create(ctx, req.WorkspaceId, userIDFrom(ctx), req.Name)
	if err != nil {
		return nil, err
	}
	return &chatv1.CreateChannelCategoryResponse{Category: presenter.ChannelCategory(out)}, nil
}

func (s *ChannelCategoryServer) UpdateChannelCategory(ctx context.Context, req *chatv1.UpdateChannelCategoryRequest) (*chatv1.UpdateChannelCategoryResponse, error) {
	out, err := s.UC.Update(ctx, req.CategoryId, userIDFrom(ctx), req.Name)
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdateChannelCategoryResponse{Category: presenter.ChannelCategory(out)}, nil
}

func (s *ChannelCategoryServer) DeleteChannelCategory(ctx context.Context, req *chatv1.DeleteChannelCategoryRequest) (*chatv1.DeleteChannelCategoryResponse, error) {
	return &chatv1.DeleteChannelCategoryResponse{}, s.UC.Delete(ctx, req.CategoryId, userIDFrom(ctx))
}

func (s *ChannelCategoryServer) ReorderChannelCategories(ctx context.Context, req *chatv1.ReorderChannelCategoriesRequest) (*chatv1.ReorderChannelCategoriesResponse, error) {
	out, err := s.UC.Reorder(ctx, req.WorkspaceId, userIDFrom(ctx), req.CategoryIds)
	if err != nil {
		return nil, err
	}
	return &chatv1.ReorderChannelCategoriesResponse{Categories: presenter.ConvertAll(out, presenter.ChannelCategory)}, nil
}

func (s *ChannelCategoryServer) SetChannelCategory(ctx context.Context, req *chatv1.SetChannelCategoryRequest) (*chatv1.SetChannelCategoryResponse, error) {
	return &chatv1.SetChannelCategoryResponse{}, s.UC.SetChannel(ctx, req.ChannelId, userIDFrom(ctx), req.CategoryId)
}
