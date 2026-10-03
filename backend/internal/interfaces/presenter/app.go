package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	appuc "github.com/newt239/chat/internal/usecase/app"
)

var AppPermissions = map[entity.AppPermission]chatv1.AppPermission{
	entity.AppPermissionPostJoinedChannels: chatv1.AppPermission_APP_PERMISSION_POST_JOINED_CHANNELS,
	entity.AppPermissionPostPublicChannels: chatv1.AppPermission_APP_PERMISSION_POST_PUBLIC_CHANNELS,
	entity.AppPermissionPostThreadReplies:  chatv1.AppPermission_APP_PERMISSION_POST_THREAD_REPLIES,
	entity.AppPermissionOutgoingWebhook:    chatv1.AppPermission_APP_PERMISSION_OUTGOING_WEBHOOK,
}

func App(a appuc.Output) *chatv1.App {
	return &chatv1.App{
		Id:               a.ID,
		WorkspaceId:      a.WorkspaceID,
		Name:             a.Name,
		Description:      a.Description,
		AvatarUrl:        a.AvatarURL,
		Permissions:      ConvertAll(a.Permissions, func(p entity.AppPermission) chatv1.AppPermission { return AppPermissions[p] }),
		DefaultChannelId: a.DefaultChannelID,
		OutgoingUrl:      a.OutgoingURL,
		OutgoingSecret:   a.OutgoingSecret,
		IsOfficial:       a.IsOfficial,
		BotUserId:        a.BotUserID,
		CreatedBy:        UserSummary(a.CreatedBy),
		CreatedAt:        timestamppb.New(a.CreatedAt),
		LastUsedAt:       optionalTimestamp(a.LastUsedAt),
		CanManage:        a.CanManage,
	}
}
