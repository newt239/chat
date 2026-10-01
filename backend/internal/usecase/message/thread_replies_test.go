package message

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

// fakeThreadRepo は返信を DB と同じように範囲・向き・件数で絞り込みます
type fakeThreadRepo struct {
	domainrepository.MessageRepository
	replies []*entity.Message
}

func (r *fakeThreadRepo) FindByID(_ context.Context, id string) (*entity.Message, error) {
	for _, m := range r.replies {
		if m.ID == id {
			return m, nil
		}
	}
	return nil, nil
}

func (r *fakeThreadRepo) FindThreadReplies(_ context.Context, _ string, limit int, since, until *time.Time, ascending bool) ([]*entity.Message, error) {
	var found []*entity.Message
	for _, m := range r.replies {
		if (since == nil || m.CreatedAt.After(*since)) && (until == nil || m.CreatedAt.Before(*until)) {
			found = append(found, m)
		}
	}
	if !ascending {
		slices.Reverse(found)
	}
	if limit > 0 && len(found) > limit {
		found = found[:limit]
	}
	return found, nil
}

func replyIDs(replies []*entity.Message) []string {
	ids := make([]string, len(replies))
	for i, m := range replies {
		ids[i] = m.ID
	}
	return ids
}

func TestFetchThreadReplies(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	parent := "parent"
	repo := &fakeThreadRepo{}
	for i := range 10 {
		repo.replies = append(repo.replies, &entity.Message{ID: fmt.Sprintf("r%d", i), ParentID: &parent, CreatedAt: base.Add(time.Duration(i) * time.Minute)})
	}
	lister := &MessageLister{messageRepo: repo}
	at := func(i int) *time.Time { return new(base.Add(time.Duration(i) * time.Minute)) }

	tests := []struct {
		name         string
		input        GetThreadRepliesInput
		want         []string
		wantHasMore  bool
		wantHasNewer bool
	}{
		{name: "指定がなければ最新を古い順に返す", input: GetThreadRepliesInput{Limit: 3}, want: []string{"r7", "r8", "r9"}, wantHasMore: true},
		{name: "until より前を返す", input: GetThreadRepliesInput{Limit: 3, Until: at(2)}, want: []string{"r0", "r1"}},
		{name: "since より後を返す", input: GetThreadRepliesInput{Limit: 3, Since: at(5)}, want: []string{"r6", "r7", "r8"}, wantHasNewer: true},
		{name: "指定した返信の前後を返す", input: GetThreadRepliesInput{Limit: 2, AroundReplyID: new("r5")}, want: []string{"r3", "r4", "r5", "r6"}, wantHasMore: true, wantHasNewer: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.MessageID = parent
			got, hasMore, hasNewer, err := lister.fetchThreadReplies(context.Background(), tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(replyIDs(got), tt.want) || hasMore != tt.wantHasMore || hasNewer != tt.wantHasNewer {
				t.Fatalf("got=%v hasMore=%t hasNewer=%t", replyIDs(got), hasMore, hasNewer)
			}
		})
	}
}
