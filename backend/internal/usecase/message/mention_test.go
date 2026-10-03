package message

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type mentionWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
	isMember bool
}

func (r *mentionWorkspaceRepo) FindMember(_ context.Context, _ string, _ string) (*entity.WorkspaceMember, error) {
	if !r.isMember {
		return nil, nil
	}
	return &entity.WorkspaceMember{}, nil
}

type mentionMessageRepo struct {
	builderMessageRepo
	found []*entity.Message
	input *domainrepository.FindMentionsInput
}

func (r *mentionMessageRepo) FindMentions(_ context.Context, input domainrepository.FindMentionsInput) ([]*entity.Message, error) {
	r.input = &input
	return r.found[:min(input.Limit, len(r.found))], nil
}

func newMentionLister(isMember bool, messages []*entity.Message) (*Interactor, *mentionMessageRepo) {
	messageRepo := &mentionMessageRepo{found: messages}
	builder := NewMessageOutputBuilder(messageRepo, &builderUserRepo{}, nil, builderMentionRepo{}, &builderLinkRepo{}, &builderAttachmentRepo{}, &builderPinRepo{}, builderPollRepo{}, nil)
	return &Interactor{workspaceRepo: &mentionWorkspaceRepo{isMember: isMember}, messageRepo: messageRepo, outputBuilder: builder}, messageRepo
}

func mentionMessages(n int) []*entity.Message {
	base := time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)
	result := make([]*entity.Message, 0, n)
	for i := range n {
		result = append(result, &entity.Message{ID: fmt.Sprintf("m%d", i), Body: "@alice", CreatedAt: base.Add(-time.Duration(i) * time.Minute)})
	}
	return result
}

func TestListMentionsRequiresMembership(t *testing.T) {
	lister, messageRepo := newMentionLister(false, mentionMessages(1))

	_, err := lister.ListMentions(context.Background(), domainrepository.FindMentionsInput{WorkspaceID: "ws", UserID: "u1"})

	if !errors.Is(err, domerr.ErrUnauthorized) {
		t.Fatalf("メンバー以外が拒否されていません: %v", err)
	}
	if messageRepo.input != nil {
		t.Error("メンバー以外でメンションが取得されました")
	}
}

func TestListMentionsPagination(t *testing.T) {
	tests := []struct {
		name       string
		total      int
		limit      int
		wantCount  int
		wantCursor *domainrepository.MessageCursor
	}{
		{name: "次のページがある", total: 3, limit: 2, wantCount: 2, wantCursor: &domainrepository.MessageCursor{CreatedAt: mentionMessages(2)[1].CreatedAt, MessageID: "m1"}},
		{name: "ちょうど最後まで", total: 2, limit: 2, wantCount: 2},
		{name: "0 は既定値を使う", total: 25, limit: 0, wantCount: defaultListLimit, wantCursor: &domainrepository.MessageCursor{CreatedAt: mentionMessages(defaultListLimit)[defaultListLimit-1].CreatedAt, MessageID: fmt.Sprintf("m%d", defaultListLimit-1)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lister, _ := newMentionLister(true, mentionMessages(tt.total))

			out, err := lister.ListMentions(context.Background(), domainrepository.FindMentionsInput{WorkspaceID: "ws", UserID: "u1", Limit: tt.limit})
			if err != nil {
				t.Fatalf("取得に失敗しました: %v", err)
			}
			if len(out.Messages) != tt.wantCount {
				t.Errorf("件数が期待と異なります: got=%d want=%d", len(out.Messages), tt.wantCount)
			}
			if (out.NextCursor == nil) != (tt.wantCursor == nil) || (out.NextCursor != nil && *out.NextCursor != *tt.wantCursor) {
				t.Errorf("カーソルが期待と異なります: got=%+v want=%+v", out.NextCursor, tt.wantCursor)
			}
		})
	}
}

func TestListMentionsPassesCursor(t *testing.T) {
	lister, messageRepo := newMentionLister(true, mentionMessages(1))
	cursor := domainrepository.MessageCursor{CreatedAt: time.Now(), MessageID: "m9"}

	if _, err := lister.ListMentions(context.Background(), domainrepository.FindMentionsInput{WorkspaceID: "ws", UserID: "u1", Cursor: &cursor}); err != nil {
		t.Fatalf("取得に失敗しました: %v", err)
	}
	got := messageRepo.input.Cursor
	if got == nil || got.CreatedAt != cursor.CreatedAt || got.MessageID != cursor.MessageID {
		t.Errorf("カーソルが渡されていません: %+v", got)
	}
}
