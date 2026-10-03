package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	pinuc "github.com/newt239/chat/internal/usecase/pin"
)

func PinnedMessage(p pinuc.PinnedMessageOutput) *chatv1.PinnedMessage {
	return &chatv1.PinnedMessage{Message: Message(p.Message), PinnedBy: p.PinnedBy, PinnedAt: timestamppb.New(p.PinnedAt)}
}

func PinEvent(channelID string, p pinuc.PinNotification) *chatv1.PinEvent {
	event := &chatv1.PinEvent{ChannelId: channelID, MessageId: p.MessageID, PinnedBy: p.PinnedBy, PinnedAt: timestamppb.New(p.PinnedAt)}
	if p.PinnedByUser != nil {
		event.PinnedByUser = UserSummary(*p.PinnedByUser)
	}
	return event
}
