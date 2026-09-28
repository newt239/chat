package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	mentionuc "github.com/newt239/chat/internal/usecase/mention"
)

func Mentions(out *mentionuc.ListMentionsOutput) *chatv1.ListMentionsResponse {
	res := &chatv1.ListMentionsResponse{Messages: ConvertAll(out.Messages, Message)}
	if out.NextCursor != nil {
		res.NextCursor = &chatv1.MentionCursor{CreatedAt: timestamppb.New(out.NextCursor.CreatedAt), MessageId: out.NextCursor.MessageID}
	}
	return res
}
