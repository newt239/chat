package presenter

import (
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	customemojiuc "github.com/newt239/chat/internal/usecase/customemoji"
)

func CustomEmoji(e customemojiuc.Output) *chatv1.CustomEmoji {
	return &chatv1.CustomEmoji{
		Id:        e.ID,
		Name:      e.Name,
		ImageUrl:  e.ImageURL,
		CreatedBy: UserSummary(e.CreatedBy),
		CanDelete: e.CanDelete,
	}
}
