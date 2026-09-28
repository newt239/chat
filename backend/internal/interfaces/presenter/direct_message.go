package presenter

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	dmuc "github.com/newt239/chat/internal/usecase/dm"
)

var directMessageTypes = map[string]chatv1.DirectMessageType{
	"dm":       chatv1.DirectMessageType_DIRECT_MESSAGE_TYPE_DM,
	"group_dm": chatv1.DirectMessageType_DIRECT_MESSAGE_TYPE_GROUP_DM,
}

// rfc3339Timestamp はユースケースが RFC3339 文字列で返す日時を変換します
func rfc3339Timestamp(value string) *timestamppb.Timestamp {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return timestamppb.New(t)
}

func DirectMessage(dm *dmuc.DMOutput) *chatv1.DirectMessage {
	return &chatv1.DirectMessage{
		Id:          dm.ID,
		WorkspaceId: dm.WorkspaceID,
		Name:        dm.Name,
		Description: dm.Description,
		Type:        directMessageTypes[dm.Type],
		Members: ConvertAll(dm.Members, func(m dmuc.DMMemberOutput) *chatv1.DirectMessageMember {
			return &chatv1.DirectMessageMember{UserId: m.UserID, DisplayName: m.DisplayName, AvatarUrl: m.AvatarURL}
		}),
		CreatedAt:   rfc3339Timestamp(dm.CreatedAt),
		UpdatedAt:   rfc3339Timestamp(dm.UpdatedAt),
		IsStarred:   dm.IsStarred,
		IsMuted:     dm.IsMuted,
		UnreadCount: int32(dm.UnreadCount),
		HasMention:  dm.HasMention,
	}
}
