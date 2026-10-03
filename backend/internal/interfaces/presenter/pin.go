package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	pinuc "github.com/newt239/chat/internal/usecase/pin"
)

func PinnedMessage(m messageuc.MessageOutput) *chatv1.PinnedMessage {
	return &chatv1.PinnedMessage{Message: Message(m), PinnedBy: m.Pin.PinnedBy.ID, PinnedAt: timestamppb.New(m.Pin.PinnedAt)}
}

func PinEvent(channelID string, p pinuc.PinNotification) *chatv1.PinEvent {
	event := &chatv1.PinEvent{ChannelId: channelID, MessageId: p.MessageID, PinnedBy: p.PinnedBy, PinnedAt: timestamppb.New(p.PinnedAt)}
	if p.PinnedByUser != nil {
		event.PinnedByUser = UserSummary(*p.PinnedByUser)
	}
	return event
}
