package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type ThreadRepository interface {
	FindMetadataByMessageID(ctx context.Context, messageID string) (*entity.ThreadMetadata, error)
	FindMetadataByMessageIDs(ctx context.Context, messageIDs []string) (map[string]*entity.ThreadMetadata, error)
	CreateOrUpdateMetadata(ctx context.Context, metadata *entity.ThreadMetadata) error
	IncrementReplyCount(ctx context.Context, messageID string, replyUserID string) error
	DeleteMetadata(ctx context.Context, messageID string) error

	// 参加中スレッド一覧取得
	FindParticipatingThreads(ctx context.Context, input FindParticipatingThreadsInput) (*FindParticipatingThreadsOutput, error)

	// スレッド既読状態管理
	UpsertReadState(ctx context.Context, userID, threadID string, lastReadAt time.Time) error
	GetReadState(ctx context.Context, userID, threadID string) (*time.Time, error)

	// スレッドフォロー管理
	FollowThread(ctx context.Context, userID, threadID string) error
	UnfollowThread(ctx context.Context, userID, threadID string) error
	IsFollowing(ctx context.Context, userID, threadID string) (bool, error)
}

type FindParticipatingThreadsInput struct {
	WorkspaceID         string
	UserID              string
	CursorLastActivityAt *time.Time
	CursorThreadID      *string
	Limit               int
}

type ParticipatingThread struct {
	ThreadID        string
	ChannelID       *string
	FirstMessage    *entity.Message
	ReplyCount      int
	LastActivityAt  time.Time
	UnreadCount     int
}

type FindParticipatingThreadsOutput struct {
	Items              []ParticipatingThread
	NextCursor         *ThreadCursor
}

type ThreadCursor struct {
	LastActivityAt time.Time
	ThreadID       string
}
