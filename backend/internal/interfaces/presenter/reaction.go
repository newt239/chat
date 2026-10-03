package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	reactionuc "github.com/newt239/chat/internal/usecase/reaction"
)

func Reaction(r reactionuc.ReactionOutput) *chatv1.Reaction {
	return &chatv1.Reaction{MessageId: r.MessageID, User: UserSummary(r.User), Emoji: r.Emoji, CreatedAt: timestamppb.New(r.CreatedAt)}
}

func ReactionEvent(channelID string, r reactionuc.ReactionNotification) *chatv1.ReactionEvent {
	event := &chatv1.ReactionEvent{ChannelId: channelID, MessageId: r.MessageID, UserId: r.UserID, Emoji: r.Emoji}
	if r.User != nil {
		event.User = UserSummary(*r.User)
		event.CreatedAt = timestamppb.New(r.CreatedAt)
	}
	return event
}
