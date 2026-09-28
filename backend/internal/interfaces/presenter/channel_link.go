package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	channellinkuc "github.com/newt239/chat/internal/usecase/channellink"
)

func ChannelLink(l channellinkuc.LinkOutput) *chatv1.ChannelLink {
	return &chatv1.ChannelLink{
		Id:        l.ID,
		ChannelId: l.ChannelID,
		Title:     l.Title,
		Url:       l.URL,
		Position:  int32(l.Position),
		CreatedBy: l.CreatedBy,
		CreatedAt: timestamppb.New(l.CreatedAt),
		UpdatedAt: timestamppb.New(l.UpdatedAt),
	}
}
