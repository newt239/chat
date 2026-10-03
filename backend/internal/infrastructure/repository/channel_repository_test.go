package repository

import (
	"context"
	"errors"
	"reflect"
	"testing"

	domerr "github.com/newt239/chat/internal/domain/errors"

	"github.com/newt239/chat/internal/domain/entity"
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

func TestFindOrCreateDMReusesChannelByKey(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	repo := NewChannelRepository(client)
	ctx := context.Background()
	alice, bob := f.alice.ID.String(), f.bob.ID.String()

	dm, err := repo.FindOrCreateDM(ctx, f.workspaceID, alice, bob)
	if err != nil {
		t.Fatalf("DM の作成に失敗しました: %v", err)
	}
	again, err := repo.FindOrCreateDM(ctx, f.workspaceID, bob, alice)
	if err != nil || again.ID != dm.ID {
		t.Fatalf("相手から開いても同じ DM を返すことを期待しましたが %v, %v でした", again, err)
	}

	group, err := repo.FindOrCreateGroupDM(ctx, f.workspaceID, alice, []string{alice, bob}, "")
	if err != nil || group.ID == dm.ID {
		t.Fatalf("グループ DM は 1:1 の DM と別に作ることを期待しましたが %v, %v でした", group, err)
	}
	same, err := repo.FindOrCreateGroupDM(ctx, f.workspaceID, bob, []string{bob, alice}, "別名")
	if err != nil || same.ID != group.ID {
		t.Fatalf("メンバーが同じグループ DM を返すことを期待しましたが %v, %v でした", same, err)
	}
	members, err := NewChannelMemberRepository(client).FindMembersByChannelIDs(ctx, []string{dm.ID, group.ID})
	if err != nil || len(members) != 4 {
		t.Fatalf("DM とグループ DM に 2 人ずつ参加していることを期待しましたが %d 人, %v でした", len(members), err)
	}
}

func TestAddMemberReportsDuplicate(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	repo := NewChannelMemberRepository(client)
	ctx := context.Background()
	member := &entity.ChannelMember{ChannelID: f.channels["general"].ID.String(), UserID: f.alice.ID.String(), Role: entity.ChannelRoleMember}

	if err := repo.AddMember(ctx, member); !errors.Is(err, domerr.ErrAlreadyMember) {
		t.Fatalf("参加済みのメンバーは ErrAlreadyMember を返すことを期待しましたが %v でした", err)
	}
	members, err := repo.FindMembersByChannelIDs(ctx, []string{member.ChannelID})
	if err != nil || len(members) != 2 {
		t.Fatalf("メンバーが重複していないことを期待しましたが %d 人, %v でした", len(members), err)
	}
}
