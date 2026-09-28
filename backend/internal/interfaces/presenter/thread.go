package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	threaduc "github.com/newt239/chat/internal/usecase/thread"
)

func ParticipatingThreads(out *threaduc.ListParticipatingThreadsOutput) *chatv1.ListParticipatingThreadsResponse {
	res := &chatv1.ListParticipatingThreadsResponse{
		Threads: ConvertAll(out.Items, func(t threaduc.ParticipatingThreadOutput) *chatv1.ParticipatingThread {
			thread := &chatv1.ParticipatingThread{
				ThreadId:       t.ThreadID,
				ChannelId:      t.ChannelID,
				ReplyCount:     int32(t.ReplyCount),
				LastActivityAt: timestamppb.New(t.LastActivityAt),
				UnreadCount:    int32(t.UnreadCount),
				LatestReplies:  ConvertAll(t.LatestReplies, Message),
			}
			if t.FirstMessage != nil {
				thread.FirstMessage = Message(*t.FirstMessage)
			}
			return thread
		}),
	}
	if out.NextCursor != nil {
		res.NextCursor = &chatv1.ThreadCursor{LastActivityAt: timestamppb.New(out.NextCursor.LastActivityAt), ThreadId: out.NextCursor.ThreadID}
	}
	return res
}
