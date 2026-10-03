package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	dmuc "github.com/newt239/chat/internal/usecase/dm"
)

var directMessageTypes = map[entity.ChannelType]chatv1.DirectMessageType{
	entity.ChannelTypeDM:      chatv1.DirectMessageType_DIRECT_MESSAGE_TYPE_DM,
	entity.ChannelTypeGroupDM: chatv1.DirectMessageType_DIRECT_MESSAGE_TYPE_GROUP_DM,
}

func DirectMessage(dm *dmuc.DMOutput) *chatv1.DirectMessage {
	return &chatv1.DirectMessage{
		Id:          dm.ID,
		WorkspaceId: dm.WorkspaceID,
		Name:        dm.Name,
		Description: dm.Description,
		Type:        directMessageTypes[dm.Type],
		Members: ConvertAll(dm.Members, func(u *entity.User) *chatv1.DirectMessageMember {
			return &chatv1.DirectMessageMember{UserId: u.ID, DisplayName: u.DisplayName, AvatarUrl: u.AvatarURL}
		}),
		CreatedAt:   timestamppb.New(dm.CreatedAt),
		UpdatedAt:   timestamppb.New(dm.UpdatedAt),
		IsStarred:   dm.IsStarred,
		IsMuted:     dm.IsMuted,
		UnreadCount: int32(dm.UnreadCount),
	}
}
