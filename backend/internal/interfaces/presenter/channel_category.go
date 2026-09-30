package presenter

import (
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	channelcategoryuc "github.com/newt239/chat/internal/usecase/channelcategory"
)

func ChannelCategory(c channelcategoryuc.CategoryOutput) *chatv1.ChannelCategory {
	return &chatv1.ChannelCategory{
		Id:         c.ID,
		Name:       c.Name,
		Position:   int32(c.Position),
		ChannelIds: c.ChannelIDs,
	}
}
