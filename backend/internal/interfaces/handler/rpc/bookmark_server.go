package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	bookmarkuc "github.com/newt239/chat/internal/usecase/bookmark"
)

type BookmarkServer struct {
	UC bookmarkuc.BookmarkUseCase
}

func (s *BookmarkServer) ListBookmarks(ctx context.Context, _ *chatv1.ListBookmarksRequest) (*chatv1.ListBookmarksResponse, error) {
	out, err := s.UC.ListBookmarks(ctx, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListBookmarksResponse{Bookmarks: presenter.ConvertAll(out.Bookmarks, presenter.Bookmark)}, nil
}

func (s *BookmarkServer) AddBookmark(ctx context.Context, req *chatv1.AddBookmarkRequest) (*chatv1.AddBookmarkResponse, error) {
	if err := s.UC.AddBookmark(ctx, bookmarkuc.AddBookmarkInput{UserID: userIDFrom(ctx), MessageID: req.MessageId}); err != nil {
		return nil, err
	}
	return &chatv1.AddBookmarkResponse{}, nil
}

func (s *BookmarkServer) RemoveBookmark(ctx context.Context, req *chatv1.RemoveBookmarkRequest) (*chatv1.RemoveBookmarkResponse, error) {
	if err := s.UC.RemoveBookmark(ctx, bookmarkuc.RemoveBookmarkInput{UserID: userIDFrom(ctx), MessageID: req.MessageId}); err != nil {
		return nil, err
	}
	return &chatv1.RemoveBookmarkResponse{}, nil
}
