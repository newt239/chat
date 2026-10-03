package rpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	bookmarkuc "github.com/newt239/chat/internal/usecase/bookmark"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

type BookmarkServer struct {
	UC *bookmarkuc.Interactor
}

func (s *BookmarkServer) ListBookmarks(ctx context.Context, _ *chatv1.ListBookmarksRequest) (*chatv1.ListBookmarksResponse, error) {
	out, err := s.UC.ListBookmarks(ctx, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListBookmarksResponse{Bookmarks: presenter.ConvertAll(out, func(b bookmarkuc.Output) *chatv1.Bookmark {
		return &chatv1.Bookmark{Message: presenter.Message(b.Message), CreatedAt: timestamppb.New(b.CreatedAt)}
	})}, nil
}

func (s *BookmarkServer) AddBookmark(ctx context.Context, req *chatv1.AddBookmarkRequest) (*chatv1.AddBookmarkResponse, error) {
	return &chatv1.AddBookmarkResponse{}, s.UC.AddBookmark(ctx, messageuc.MessageInput{MessageID: req.MessageId, UserID: userIDFrom(ctx)})
}

func (s *BookmarkServer) RemoveBookmark(ctx context.Context, req *chatv1.RemoveBookmarkRequest) (*chatv1.RemoveBookmarkResponse, error) {
	return &chatv1.RemoveBookmarkResponse{}, s.UC.RemoveBookmark(ctx, messageuc.MessageInput{MessageID: req.MessageId, UserID: userIDFrom(ctx)})
}
