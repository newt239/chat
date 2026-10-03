package repository

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

func TestMembershipBatchQueries(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	ctx := context.Background()
	alice, bob := f.alice.ID.String(), f.bob.ID.String()
	general, secret := f.channels["general"].ID.String(), f.channels["secret"].ID.String()

	joined, err := NewChannelMemberRepository(client).FindJoinedChannelIDs(ctx, alice, []string{general, secret})
	if err != nil || !reflect.DeepEqual(joined, map[string]bool{general: true}) {
		t.Errorf("参加中のチャンネルが期待と異なります: %v %v", joined, err)
	}
	members, err := NewChannelMemberRepository(client).FindMemberIDsIn(ctx, secret, []string{alice, bob})
	if err != nil || !reflect.DeepEqual(members, map[string]bool{bob: true}) {
		t.Errorf("チャンネルの参加者が期待と異なります: %v %v", members, err)
	}

	workspaces := NewWorkspaceRepository(client)
	if err := workspaces.SetMemberSuspended(ctx, f.workspaceID, bob, new(time.Now())); err != nil {
		t.Fatal(err)
	}
	active, err := workspaces.FindActiveMemberIDs(ctx, f.workspaceID, []string{alice, bob})
	if err != nil || !reflect.DeepEqual(active, map[string]bool{alice: true}) {
		t.Errorf("停止中のメンバーを除いていません: %v %v", active, err)
	}
	counts, err := workspaces.CountMembersBatch(ctx, []string{f.workspaceID, "missing"})
	if err != nil || !reflect.DeepEqual(counts, map[string]int{f.workspaceID: 2}) {
		t.Errorf("メンバー数が期待と異なります: %v %v", counts, err)
	}
	memberships, err := workspaces.FindMembershipsByUserID(ctx, bob)
	if err != nil || len(memberships) != 0 {
		t.Errorf("停止中のワークスペースを返しています: %v %v", memberships, err)
	}

	mutes := NewChannelMuteRepository(client)
	for range 2 {
		if err := mutes.SetMuted(ctx, alice, general, true); err != nil {
			t.Fatalf("ミュートを繰り返すと失敗しました: %v", err)
		}
	}
	muted, err := mutes.FindMutedUserIDs(ctx, general, []string{alice, bob})
	if err != nil || !reflect.DeepEqual(muted, map[string]bool{alice: true}) {
		t.Errorf("ミュートしているユーザーが期待と異なります: %v %v", muted, err)
	}
}

func TestFindLatestSessionsByUserIDs(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	ctx := context.Background()
	repo := NewSessionRepository(client)
	alice := f.alice.ID.String()

	for _, hash := range []string{"old-" + alice, "new-" + alice} {
		if err := repo.Create(ctx, &entity.Session{UserID: alice, RefreshTokenHash: hash, ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond)
	}
	latest, err := repo.FindLatestByUserIDs(ctx, []string{alice, f.bob.ID.String()})
	if err != nil || len(latest) != 1 || latest[alice].RefreshTokenHash != "new-"+alice {
		t.Errorf("最後のセッションが期待と異なります: %+v %v", latest, err)
	}
}

func TestAttachToMessageUpdatesAllAtOnce(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	ctx := context.Background()
	repo := NewAttachmentRepository(client)

	var ids []string
	for range 2 {
		a := client.Attachment.Create().SetFileName("f").SetMimeType("text/plain").SetSizeBytes(1).SetStorageKey("k").
			SetUploader(f.alice).SetChannel(f.channels["general"]).SaveX(ctx)
		ids = append(ids, a.ID.String())
	}
	if err := repo.AttachToMessage(ctx, ids, f.messages["mention"].ID.String()); err != nil {
		t.Fatal(err)
	}
	attached, err := repo.FindByMessageIDs(ctx, []string{f.messages["mention"].ID.String()})
	if err != nil || len(attached[f.messages["mention"].ID.String()]) != 2 {
		t.Errorf("添付が付いていません: %v %v", attached, err)
	}
}
