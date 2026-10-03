package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type ThreadMetadata struct {
	ReplyCount      int
	LastReplyAt     *time.Time
	LastReplyUserID *string
}

type ThreadRepository interface {
	// CalculateMetadataByMessageIDs は返信のないメッセージも含め、指定した全メッセージのメタデータを返します
	CalculateMetadataByMessageIDs(ctx context.Context, messageIDs []string) (map[string]*ThreadMetadata, error)
	FindParticipatingThreads(ctx context.Context, input FindParticipatingThreadsInput) (*FindParticipatingThreadsOutput, error)

	UpsertReadState(ctx context.Context, userID, threadID string, lastReadAt time.Time) error

	// SetFollowing は既にその状態でも成功します
	SetFollowing(ctx context.Context, userID, threadID string, following bool) error
	// FindFollowedThreadIDs は threadIDs のうち userID がフォローしているものを返します
	FindFollowedThreadIDs(ctx context.Context, userID string, threadIDs []string) (map[string]bool, error)
	FindFollowerIDs(ctx context.Context, threadID string) ([]string, error)
}

type FindParticipatingThreadsInput struct {
	WorkspaceID string
	UserID      string
	Cursor      *ThreadCursor
	Limit       int
}

type ParticipatingThread struct {
	ThreadID       string
	FirstMessage   *entity.Message
	LatestReplies  []*entity.Message
	ReplyCount     int
	LastActivityAt time.Time
	UnreadCount    int
}

type FindParticipatingThreadsOutput struct {
	Items      []ParticipatingThread
	NextCursor *ThreadCursor
}

type ThreadCursor struct {
	LastActivityAt time.Time
	ThreadID       string
}
