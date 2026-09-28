package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	webhookuc "github.com/newt239/chat/internal/usecase/webhook"
)

func Webhook(w webhookuc.Output) *chatv1.Webhook {
	return &chatv1.Webhook{
		Id:         w.ID,
		ChannelId:  w.ChannelID,
		Name:       w.Name,
		AvatarUrl:  w.AvatarURL,
		CreatedBy:  UserSummary(w.CreatedBy),
		CreatedAt:  timestamppb.New(w.CreatedAt),
		LastUsedAt: optionalTimestamp(w.LastUsedAt),
		CanManage:  w.CanManage,
	}
}
