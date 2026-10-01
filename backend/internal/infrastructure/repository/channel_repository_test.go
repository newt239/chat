package repository

import (
	"context"
	"reflect"
	"testing"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

func TestSearchBrowsableChannels(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	repo := NewChannelRepository(client)
	alice := f.alice.ID.String()

	tests := []struct {
		name      string
		filter    domainrepository.BrowsableChannelFilter
		wantNames []string
		wantTotal int
	}{
		{
			// 参加していない非公開の secret は含めない。同じ人数なら名前順にする
			name:      "メンバーの多い順",
			filter:    domainrepository.BrowsableChannelFilter{Sort: domainrepository.BrowsableChannelSortMemberCount, Limit: 10},
			wantNames: []string{"dev", "general", "dev/web"},
			wantTotal: 3,
		},
		{
			name:      "未参加だけ",
			filter:    domainrepository.BrowsableChannelFilter{Membership: domainrepository.BrowsableChannelMembershipNotJoined, Limit: 10},
			wantNames: []string{"dev/web"},
			wantTotal: 1,
		},
		{
			name:      "参加中をキーワードで絞り込む",
			filter:    domainrepository.BrowsableChannelFilter{Query: "GEN", Membership: domainrepository.BrowsableChannelMembershipJoined, Limit: 10},
			wantNames: []string{"general"},
			wantTotal: 1,
		},
		{
			name:      "2 ページ目",
			filter:    domainrepository.BrowsableChannelFilter{Limit: 1, Offset: 1},
			wantNames: []string{"dev/web"},
			wantTotal: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channels, total, err := repo.SearchBrowsableChannels(context.Background(), f.workspaceID, alice, tt.filter)
			if err != nil {
				t.Fatalf("検索に失敗しました: %v", err)
			}
			names := make([]string, len(channels))
			for i, ch := range channels {
				names[i] = ch.Name
			}
			if !reflect.DeepEqual(names, tt.wantNames) || total != tt.wantTotal {
				t.Errorf("got=%v (%d) want=%v (%d)", names, total, tt.wantNames, tt.wantTotal)
			}
		})
	}
}
