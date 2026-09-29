package search

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
	isMember bool
}

func (r *stubWorkspaceRepo) FindByID(_ context.Context, id string) (*entity.Workspace, error) {
	return &entity.Workspace{ID: id}, nil
}

func (r *stubWorkspaceRepo) FindMember(_ context.Context, workspaceID string, userID string) (*entity.WorkspaceMember, error) {
	if !r.isMember {
		return nil, nil
	}
	return &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: userID}, nil
}

func (r *stubWorkspaceRepo) SearchMembers(_ context.Context, _ string, _ string, _ int, _ int) ([]*entity.WorkspaceMember, int, error) {
	return []*entity.WorkspaceMember{}, 0, nil
}

type stubChannelRepo struct {
	domainrepository.ChannelRepository
	channels      []*entity.Channel
	searchedQuery *string
}

func (r *stubChannelRepo) FindByWorkspaceID(_ context.Context, _ string) ([]*entity.Channel, error) {
	return r.channels, nil
}

func (r *stubChannelRepo) SearchAccessibleChannels(_ context.Context, _ string, _ string, query string, _ int, _ int) ([]*entity.Channel, int, error) {
	r.searchedQuery = &query
	return []*entity.Channel{}, 0, nil
}

type stubMessageRepo struct {
	domainrepository.MessageRepository
	scope    domainrepository.MessageSearchScope
	messages []*entity.Message
}

func (r *stubMessageRepo) FindSearchScope(_ context.Context, _ string, _ string) (*domainrepository.MessageSearchScope, error) {
	return &r.scope, nil
}

func (r *stubMessageRepo) FindByIDs(_ context.Context, _ []string) ([]*entity.Message, error) {
	return r.messages, nil
}

type stubSearchIndex struct {
	domainrepository.MessageSearchIndex
	criteria *domainrepository.MessageSearchCriteria
}

func (i *stubSearchIndex) Search(_ context.Context, c domainrepository.MessageSearchCriteria) (*domainrepository.MessageSearchResult, error) {
	i.criteria = &c
	return &domainrepository.MessageSearchResult{MessageIDs: []string{}}, nil
}

func newSearcher(channels []*entity.Channel, scope domainrepository.MessageSearchScope) (*WorkspaceSearcher, *stubChannelRepo, *stubSearchIndex) {
	channelRepo := &stubChannelRepo{channels: channels}
	index := &stubSearchIndex{}
	return NewWorkspaceSearcher(&stubWorkspaceRepo{isMember: true}, channelRepo, &stubMessageRepo{scope: scope}, index, nil, nil, nil), channelRepo, index
}

func TestSearchWorkspaceValidation(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-time.Hour)
	tests := []struct {
		name    string
		input   WorkspaceSearchInput
		wantErr error
	}{
		{name: "キーワードも条件もない", input: WorkspaceSearchInput{Query: "  　 "}, wantErr: ErrInvalidQuery},
		{name: "返信を除くだけでは条件とみなさない", input: WorkspaceSearchInput{Filter: MessageFilter{ExcludeReplies: true}}, wantErr: ErrInvalidQuery},
		{name: "期間の開始が終了以降", input: WorkspaceSearchInput{Query: "a", Filter: MessageFilter{After: &now, Before: &earlier}}, wantErr: ErrInvalidDateRange},
		{name: "条件だけで検索できる", input: WorkspaceSearchInput{Filter: MessageFilter{PinnedOnly: true}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searcher, _, _ := newSearcher(nil, domainrepository.MessageSearchScope{ViewableChannelIDs: []string{"c"}})
			_, err := searcher.SearchWorkspace(context.Background(), tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("エラーが期待と異なります: got=%v want=%v", err, tt.wantErr)
			}
		})
	}
}

func TestSearchWorkspaceRequiresMembership(t *testing.T) {
	searcher := NewWorkspaceSearcher(&stubWorkspaceRepo{}, &stubChannelRepo{}, &stubMessageRepo{}, &stubSearchIndex{}, nil, nil, nil)
	_, err := searcher.SearchWorkspace(context.Background(), WorkspaceSearchInput{Query: "a"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("メンバー以外の検索が拒否されていません: %v", err)
	}
}

func TestSearchWorkspaceBuildsCriteria(t *testing.T) {
	scope := domainrepository.MessageSearchScope{
		UserID:             "u1",
		ViewableChannelIDs: []string{"dev", "dev-web", "devops", "general"},
		JoinedChannelIDs:   []string{"dev"},
		GroupIDs:           []string{"g1"},
	}
	searcher, _, index := newSearcher([]*entity.Channel{
		{ID: "dev", Name: "dev"},
		{ID: "dev-web", Name: "dev/web"},
		{ID: "dev-web-a", Name: "dev/web/a"},
		{ID: "devops", Name: "devops"},
	}, scope)
	after := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	_, err := searcher.SearchWorkspace(context.Background(), WorkspaceSearchInput{
		WorkspaceID: "ws",
		RequesterID: "u1",
		Query:       "設計 Review 設計",
		Target:      SearchTargetMessages,
		Sort:        domainrepository.MessageSearchSortRelevance,
		Page:        2,
		PerPage:     100,
		Filter: MessageFilter{
			FromUserIDs:               []string{"u2"},
			ChannelIDs:                []string{"dev"},
			IncludeDescendantChannels: true,
			Has:                       []domainrepository.MessageContentKind{domainrepository.MessageContentImage},
			MentionsMe:                true,
			After:                     &after,
		},
	})
	if err != nil {
		t.Fatalf("検索に失敗しました: %v", err)
	}

	want := domainrepository.MessageSearchCriteria{
		WorkspaceID: "ws",
		Terms:       []string{"設計", "Review"},
		// 下階層のうち閲覧できない dev/web/a は除く
		ChannelIDs: []string{"dev", "dev-web"},
		AuthorIDs:  []string{"u2"},
		Has:        []domainrepository.MessageContentKind{domainrepository.MessageContentImage},
		Mention:    &scope,
		After:      &after,
		Sort:       domainrepository.MessageSearchSortRelevance,
		Page:       2,
		PerPage:    maxPerPage,
	}
	if !reflect.DeepEqual(*index.criteria, want) {
		t.Errorf("検索条件が期待と異なります:\n got=%+v\nwant=%+v", *index.criteria, want)
	}
}

func TestSearchWorkspaceWithoutKeywordSortsByNewest(t *testing.T) {
	searcher, channelRepo, index := newSearcher(nil, domainrepository.MessageSearchScope{ViewableChannelIDs: []string{"c"}})

	_, err := searcher.SearchWorkspace(context.Background(), WorkspaceSearchInput{
		Filter: MessageFilter{ThreadOnly: true},
		Sort:   domainrepository.MessageSearchSortRelevance,
	})
	if err != nil {
		t.Fatalf("検索に失敗しました: %v", err)
	}
	if index.criteria == nil || index.criteria.Sort != domainrepository.MessageSearchSortNewest {
		t.Errorf("キーワードなしの検索が新しい順になっていません: %+v", index.criteria)
	}
	if channelRepo.searchedQuery != nil {
		t.Error("キーワードなしでチャンネルが検索されました")
	}
}

func TestSearchWorkspaceSkipsIndexWithoutViewableChannels(t *testing.T) {
	searcher, _, index := newSearcher(nil, domainrepository.MessageSearchScope{ViewableChannelIDs: []string{"general"}})

	out, err := searcher.SearchWorkspace(context.Background(), WorkspaceSearchInput{Filter: MessageFilter{ChannelIDs: []string{"secret"}}})
	if err != nil {
		t.Fatalf("検索に失敗しました: %v", err)
	}
	if index.criteria != nil || out.Messages.Total != 0 {
		t.Errorf("閲覧できないチャンネルだけを指定したのに検索されました: %+v", index.criteria)
	}
}

func TestFindLiveMessagesKeepsIndexOrder(t *testing.T) {
	deletedAt := time.Now()
	searcher := NewWorkspaceSearcher(nil, nil, &stubMessageRepo{messages: []*entity.Message{
		{ID: "a"}, {ID: "b", DeletedAt: &deletedAt}, {ID: "c"},
	}}, nil, nil, nil, nil)

	messages, err := searcher.findLiveMessages(context.Background(), []string{"c", "missing", "b", "a"})
	if err != nil {
		t.Fatalf("取得に失敗しました: %v", err)
	}
	got := []string{}
	for _, m := range messages {
		got = append(got, m.ID)
	}
	if !reflect.DeepEqual(got, []string{"c", "a"}) {
		t.Errorf("インデックスの順序で削除済み・存在しないものを除けていません: %v", got)
	}
}

func TestHighlightRanges(t *testing.T) {
	tests := []struct {
		name  string
		body  string
		terms []string
		want  []TextRange
	}{
		{name: "大文字小文字を区別しない", body: "Go と go", terms: []string{"GO"}, want: []TextRange{{0, 2}, {5, 7}}},
		{name: "日本語", body: "設計レビューの設計", terms: []string{"設計"}, want: []TextRange{{0, 2}, {7, 9}}},
		{name: "サロゲートペアは UTF-16 で数える", body: "🎉リリース", terms: []string{"リリース"}, want: []TextRange{{2, 6}}},
		{name: "重なる範囲はまとめる", body: "abcd", terms: []string{"abc", "bcd"}, want: []TextRange{{0, 4}}},
		{name: "一致しない", body: "abc", terms: []string{"x"}, want: []TextRange{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := highlightRanges(tt.body, tt.terms); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got=%v want=%v", got, tt.want)
			}
		})
	}
}
