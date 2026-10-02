package repository

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/newt239/chat/ent/channelreadstate"

	"github.com/newt239/chat/internal/domain/entity"
)

func TestUnreadCountBatch(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	ctx := context.Background()
	repo := NewReadStateRepository(client)
	general, dev := f.channels["general"].ID.String(), f.channels["dev"].ID.String()
	alice := f.alice.ID.String()
	client.ChannelReadState.Create().SetChannel(f.channels["general"]).SetUser(f.alice).
		SetLastReadAt(time.Date(2026, 9, 1, 0, 2, 0, 0, time.UTC)).SaveX(ctx)

	unread, err := repo.GetUnreadCountBatch(ctx, []string{general, dev}, alice)
	if err != nil {
		t.Fatalf("未読数の取得に失敗しました: %v", err)
	}
	// general は既読位置より後の削除されていない 5 件、dev は既読位置がないので全件
	if want := map[string]int{general: 5, dev: 2}; !reflect.DeepEqual(unread, want) {
		t.Errorf("未読数が期待と異なります: got=%v want=%v", unread, want)
	}

	mentions, err := repo.GetUnreadMentionCountBatch(ctx, []string{general, dev}, alice)
	if err != nil {
		t.Fatalf("未読メンション数の取得に失敗しました: %v", err)
	}
	// 本人宛てとグループ宛てだけを数え、@channel は含めない
	if want := map[string]int{general: 2}; !reflect.DeepEqual(mentions, want) {
		t.Errorf("未読メンション数が期待と異なります: got=%v want=%v", mentions, want)
	}

	if err := repo.Upsert(ctx, &entity.ChannelReadState{ChannelID: general, UserID: alice, LastReadAt: time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)}); err != nil {
		t.Fatalf("既読位置の更新に失敗しました: %v", err)
	}
	if count, _ := repo.GetUnreadCount(ctx, general, alice); count != 0 {
		t.Errorf("既読位置を更新しても未読が残っています: %d", count)
	}
}

func TestCalculateMetadataByMessageIDs(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	ctx := context.Background()
	client.UserThreadFollow.Create().SetUser(f.bob).SetThread(f.messages["mention"]).SaveX(ctx)
	thread, single := f.messages["mention"].ID.String(), f.messages["link"].ID.String()

	metadata, err := NewThreadRepository(client).CalculateMetadataByMessageIDs(ctx, []string{thread, single})
	if err != nil {
		t.Fatalf("取得に失敗しました: %v", err)
	}
	got := metadata[thread]
	if got.ReplyCount != 1 || *got.LastReplyUserID != f.alice.ID.String() || !got.LastReplyAt.Equal(f.messages["reply"].CreatedAt) {
		t.Errorf("スレッドのメタデータが期待と異なります: %+v", got)
	}
	if metadata[single].ReplyCount != 0 || metadata[single].LastReplyAt != nil {
		t.Errorf("返信のないメッセージのメタデータが期待と異なります: %+v", metadata[single])
	}
}

func TestAdvanceReadStateBatchDoesNotMoveBackwards(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	ctx := context.Background()
	repo := NewReadStateRepository(client)
	general, dev := f.channels["general"].ID.String(), f.channels["dev"].ID.String()
	alice := f.alice.ID.String()
	later := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)

	if err := repo.Upsert(ctx, &entity.ChannelReadState{ChannelID: general, UserID: alice, LastReadAt: later}); err != nil {
		t.Fatal(err)
	}
	if err := repo.AdvanceBatch(ctx, []string{general, dev}, alice, later.Add(-time.Hour)); err != nil {
		t.Fatalf("既読位置をまとめて進められません: %v", err)
	}
	states := client.ChannelReadState.Query().Where(channelreadstate.UserID(f.alice.ID)).AllX(ctx)
	got := map[string]time.Time{}
	for _, s := range states {
		got[s.ChannelID.String()] = s.LastReadAt
	}
	if !got[general].Equal(later) || !got[dev].Equal(later.Add(-time.Hour)) {
		t.Errorf("既読位置が期待と異なります: %v", got)
	}
}
