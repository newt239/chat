package message

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
)

type stubAccess struct {
	domainservice.ChannelAccessService
}

func (stubAccess) EnsureChannelAccess(_ context.Context, id string, _ string) (*entity.Channel, error) {
	return &entity.Channel{ID: id}, nil
}

func (stubAccess) AccessibleDescendants(_ context.Context, _ *entity.Channel, _ string) ([]*entity.Channel, error) {
	return []*entity.Channel{{ID: "child"}}, nil
}

type recordingMessageRepo struct {
	domainrepository.MessageRepository
	channelIDs []string
}

func (r *recordingMessageRepo) FindByChannelIDs(_ context.Context, channelIDs []string, _ int, _ *time.Time, _ *time.Time, _ bool) ([]*entity.Message, error) {
	r.channelIDs = channelIDs
	return nil, nil
}

type recordingSystemMessageRepo struct {
	domainrepository.SystemMessageRepository
	channelIDs []string
}

func (r *recordingSystemMessageRepo) FindByChannelIDs(_ context.Context, channelIDs []string, _ int, _ *time.Time, _ *time.Time, _ bool) ([]*entity.SystemMessage, error) {
	r.channelIDs = channelIDs
	return nil, nil
}

func TestListMessagesIncludeDescendants(t *testing.T) {
	tests := []struct {
		name               string
		includeDescendants bool
		want               []string
	}{
		{name: "子孫を含める", includeDescendants: true, want: []string{"parent", "child"}},
		{name: "子孫を含めない", includeDescendants: false, want: []string{"parent"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages := &recordingMessageRepo{}
			systemMessages := &recordingSystemMessageRepo{}
			lister := &MessageLister{
				messageRepo:      messages,
				systemMsgRepo:    systemMessages,
				channelAccessSvc: stubAccess{},
				outputBuilder:    &MessageOutputBuilder{},
			}

			if _, err := lister.ListMessages(context.Background(), ListMessagesInput{ChannelID: "parent", UserID: "u", IncludeDescendants: tt.includeDescendants}); err != nil {
				t.Fatal(err)
			}

			if !slices.Equal(messages.channelIDs, tt.want) || !slices.Equal(systemMessages.channelIDs, tt.want) {
				t.Fatalf("取得対象のチャンネルが正しくありません: messages=%v system=%v", messages.channelIDs, systemMessages.channelIDs)
			}
		})
	}
}

// fakeSystemMessageRepo は作成日時の範囲・向き・件数を DB と同じように絞り込みます
type fakeSystemMessageRepo struct {
	domainrepository.SystemMessageRepository
	messages []*entity.SystemMessage
}

func (r *fakeSystemMessageRepo) FindByChannelIDs(_ context.Context, _ []string, limit int, since *time.Time, until *time.Time, ascending bool) ([]*entity.SystemMessage, error) {
	out := make([]*entity.SystemMessage, 0)
	for _, m := range r.messages {
		if (since == nil || m.CreatedAt.After(*since)) && (until == nil || m.CreatedAt.Before(*until)) {
			out = append(out, m)
		}
	}
	slices.SortFunc(out, func(a, b *entity.SystemMessage) int { return a.CreatedAt.Compare(b.CreatedAt) })
	if !ascending {
		slices.Reverse(out)
	}
	return out[:min(limit, len(out))], nil
}

func TestListMessagesAround(t *testing.T) {
	base := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	systemMessages := &fakeSystemMessageRepo{}
	// 9/1 0:00 の前後 1 時間おきに 5 件ずつ
	for i := -5; i < 5; i++ {
		systemMessages.messages = append(systemMessages.messages, &entity.SystemMessage{ID: base.Add(time.Duration(i) * time.Hour).Format("15"), CreatedAt: base.Add(time.Duration(i) * time.Hour)})
	}
	lister := &MessageLister{
		messageRepo:      &recordingMessageRepo{},
		systemMsgRepo:    systemMessages,
		channelAccessSvc: stubAccess{},
		outputBuilder:    &MessageOutputBuilder{},
	}
	ids := func(out *ListMessagesOutput) []string {
		got := make([]string, 0, len(out.Messages))
		for _, item := range out.Messages {
			got = append(got, item.SystemMessage.ID)
		}
		return got
	}

	tests := []struct {
		name         string
		input        ListMessagesInput
		want         []string
		wantHasMore  bool
		wantHasNewer bool
	}{
		{
			name:         "指定日時の前後を limit 件ずつ新しい順に返す",
			input:        ListMessagesInput{Limit: 2, Around: &base},
			want:         []string{"01", "00", "23", "22"},
			wantHasMore:  true,
			wantHasNewer: true,
		},
		{
			name:  "前後に続きがなければ has_more と has_newer は偽",
			input: ListMessagesInput{Limit: 5, Around: &base},
			want:  []string{"04", "03", "02", "01", "00", "23", "22", "21", "20", "19"},
		},
		{
			name:         "since だけを指定すると直後から limit 件を返す",
			input:        ListMessagesInput{Limit: 2, Since: &base},
			want:         []string{"02", "01"},
			wantHasNewer: true,
		},
		{
			name:        "until だけを指定すると直前から limit 件を返す",
			input:       ListMessagesInput{Limit: 2, Until: &base},
			want:        []string{"23", "22"},
			wantHasMore: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.ChannelID = "c"
			tt.input.UserID = "u"
			out, err := lister.ListMessages(context.Background(), tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if got := ids(out); !slices.Equal(got, tt.want) || out.HasMore != tt.wantHasMore || out.HasNewer != tt.wantHasNewer {
				t.Fatalf("got %v hasMore=%v hasNewer=%v, want %v hasMore=%v hasNewer=%v", got, out.HasMore, out.HasNewer, tt.want, tt.wantHasMore, tt.wantHasNewer)
			}
		})
	}
}
