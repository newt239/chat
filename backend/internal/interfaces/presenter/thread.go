package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

func ParticipatingThreads(out *messageuc.ListParticipatingThreadsOutput) *chatv1.ListParticipatingThreadsResponse {
	res := &chatv1.ListParticipatingThreadsResponse{
		Threads: ConvertAll(out.Items, func(t messageuc.ParticipatingThreadOutput) *chatv1.ParticipatingThread {
			return &chatv1.ParticipatingThread{
				ThreadId:       t.ThreadID,
				ReplyCount:     int32(t.ReplyCount),
				LastActivityAt: timestamppb.New(t.LastActivityAt),
				UnreadCount:    int32(t.UnreadCount),
				IsFollowing:    t.IsFollowing,
				LatestReplies:  ConvertAll(t.LatestReplies, Message),
				FirstMessage:   Message(t.FirstMessage),
			}
		}),
	}
	if out.NextCursor != nil {
		res.NextCursor = &chatv1.ThreadCursor{LastActivityAt: timestamppb.New(out.NextCursor.LastActivityAt), ThreadId: out.NextCursor.ThreadID}
	}
	return res
}
