package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type PollRepository interface {
	// Create は投票と選択肢を保存し、ID を設定します
	Create(ctx context.Context, poll *entity.Poll) error
	FindByID(ctx context.Context, id string) (*entity.Poll, error)
	// FindByMessageIDs はメッセージ ID ごとの投票を返します
	FindByMessageIDs(ctx context.Context, messageIDs []string) (map[string]*entity.Poll, error)
	FindVotesByPollIDs(ctx context.Context, pollIDs []string) ([]*entity.PollVote, error)
	// ReplaceVotes はユーザーのこの投票への票を optionIDs に置き換えます
	ReplaceVotes(ctx context.Context, pollID, userID string, optionIDs []string) error
	Close(ctx context.Context, id string, closedAt time.Time) error
}
