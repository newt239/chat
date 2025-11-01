package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

// ThreadMetadata はスレッドのメタデータを表します（計算結果）
type ThreadMetadata struct {
	MessageID          string
	ReplyCount         int
	LastReplyAt        *time.Time
	LastReplyUserID    *string
	ParticipantUserIDs []string
}

type ThreadRepository interface {
	// スレッドメタデータを計算して取得
	CalculateMetadataByMessageID(ctx context.Context, messageID string) (*ThreadMetadata, error)
	CalculateMetadataByMessageIDs(ctx context.Context, messageIDs []string) (map[string]*ThreadMetadata, error)

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
