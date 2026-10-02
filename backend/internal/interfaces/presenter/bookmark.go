package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	bookmarkuc "github.com/newt239/chat/internal/usecase/bookmark"
)

func Bookmark(b bookmarkuc.BookmarkWithMessageOutput) *chatv1.Bookmark {
	return &chatv1.Bookmark{UserId: b.UserID, Message: Message(b.Message), CreatedAt: timestamppb.New(b.CreatedAt)}
}
