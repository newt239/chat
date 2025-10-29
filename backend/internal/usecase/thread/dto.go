package thread

import (
	"time"

	"github.com/newt239/chat/internal/usecase/message"
)

type ListParticipatingThreadsInput struct {
	WorkspaceID          string
	UserID               string
	CursorLastActivityAt *time.Time
	CursorThreadID       *string
	Limit                int
}

type ParticipatingThreadOutput struct {
	ThreadID        string                  `json:"thread_id"`
	ChannelID       *string                 `json:"channel_id"`
	FirstMessage    *message.MessageOutput  `json:"first_message"`
	ReplyCount      int                     `json:"reply_count"`
	LastActivityAt  time.Time               `json:"last_activity_at"`
	UnreadCount     int                     `json:"unread_count"`
}

type ThreadCursorOutput struct {
	LastActivityAt time.Time `json:"last_activity_at"`
	ThreadID       string    `json:"thread_id"`
}

type ListParticipatingThreadsOutput struct {
	Items      []ParticipatingThreadOutput `json:"items"`
	NextCursor *ThreadCursorOutput         `json:"next_cursor"`
}

type MarkThreadReadInput struct {
	UserID   string
	ThreadID string
}
