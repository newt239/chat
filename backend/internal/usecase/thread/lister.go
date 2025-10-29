package thread

import (
	"context"
	"fmt"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/usecase/message"
)

type ThreadLister struct {
	threadRepo domainrepository.ThreadRepository
}

func NewThreadLister(
	threadRepo domainrepository.ThreadRepository,
) *ThreadLister {
	return &ThreadLister{
		threadRepo: threadRepo,
	}
}

func (l *ThreadLister) ListParticipatingThreads(ctx context.Context, input ListParticipatingThreadsInput) (*ListParticipatingThreadsOutput, error) {
	// デフォルトリミット設定
	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	repoInput := domainrepository.FindParticipatingThreadsInput{
		WorkspaceID:          input.WorkspaceID,
		UserID:               input.UserID,
		CursorLastActivityAt: input.CursorLastActivityAt,
		CursorThreadID:       input.CursorThreadID,
		Limit:                limit,
	}

	result, err := l.threadRepo.FindParticipatingThreads(ctx, repoInput)
	if err != nil {
		return nil, fmt.Errorf("failed to find participating threads: %w", err)
	}

	// エンティティをDTOに変換
	items := make([]ParticipatingThreadOutput, 0, len(result.Items))
	for _, item := range result.Items {
		var messageOutput *message.MessageOutput
		if item.FirstMessage != nil {
			// 最小限のメッセージ情報を構築
			msg := &message.MessageOutput{
				ID:          item.FirstMessage.ID,
				ChannelID:   item.FirstMessage.ChannelID,
				UserID:      item.FirstMessage.UserID,
				ParentID:    item.FirstMessage.ParentID,
				Body:        item.FirstMessage.Body,
				CreatedAt:   item.FirstMessage.CreatedAt,
				EditedAt:    item.FirstMessage.EditedAt,
				DeletedAt:   item.FirstMessage.DeletedAt,
				IsDeleted:   item.FirstMessage.DeletedAt != nil,
				User:        message.UserInfo{ID: item.FirstMessage.UserID, DisplayName: "", AvatarURL: nil},
				Mentions:    []message.UserMention{},
				Groups:      []message.GroupMention{},
				Links:       []message.LinkInfo{},
				Reactions:   []message.ReactionInfo{},
				Attachments: []message.AttachmentInfo{},
			}
			messageOutput = msg
		}

		items = append(items, ParticipatingThreadOutput{
			ThreadID:       item.ThreadID,
			ChannelID:      item.ChannelID,
			FirstMessage:   messageOutput,
			ReplyCount:     item.ReplyCount,
			LastActivityAt: item.LastActivityAt,
			UnreadCount:    item.UnreadCount,
		})
	}

	var nextCursor *ThreadCursorOutput
	if result.NextCursor != nil {
		nextCursor = &ThreadCursorOutput{
			LastActivityAt: result.NextCursor.LastActivityAt,
			ThreadID:       result.NextCursor.ThreadID,
		}
	}

	return &ListParticipatingThreadsOutput{
		Items:      items,
		NextCursor: nextCursor,
	}, nil
}
