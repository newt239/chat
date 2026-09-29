package repository

import (
	"context"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

func TestScheduledMessageClaimDueSkipsLockedRows(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	repo := NewScheduledMessageRepository(client)
	ctx := context.Background()
	now := time.Now()

	schedule := func(at time.Time) *entity.ScheduledMessage {
		t.Helper()
		m := &entity.ScheduledMessage{UserID: f.alice.ID.String(), ChannelID: f.channels["general"].ID.String(), Body: "予約", ScheduledAt: at}
		if err := repo.Create(ctx, m); err != nil {
			t.Fatalf("予約を作れません: %v", err)
		}
		return m
	}
	due := schedule(now.Add(-time.Minute))
	future := schedule(now.Add(time.Hour))

	// 1 つ目のトランザクションが取り出して確定させる前は、他のワーカーには見えない
	err := transaction.NewTransactionManager(client).Do(ctx, func(txCtx context.Context) error {
		claimed, err := repo.ClaimDue(txCtx, now, 10)
		if err != nil {
			return err
		}
		if len(claimed) != 1 || claimed[0].ID != due.ID || claimed[0].Status != entity.ScheduledMessageSending {
			t.Errorf("期限の来た予約だけを送信中にしていません: %+v", claimed)
		}
		concurrent, err := repo.ClaimDue(ctx, now, 10)
		if err != nil {
			return err
		}
		if len(concurrent) != 0 {
			t.Errorf("ロック中の予約を二重に取り出しています: %+v", concurrent)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("取り出しに失敗しました: %v", err)
	}

	if again, _ := repo.ClaimDue(ctx, now, 10); len(again) != 0 {
		t.Errorf("送信中の予約をもう一度取り出しています: %+v", again)
	}
	if claimed, _ := repo.Claim(ctx, due.ID); claimed != nil {
		t.Errorf("送信中の予約を今すぐ送信で取り出しています: %+v", claimed)
	}
	claimed, err := repo.Claim(ctx, future.ID)
	if err != nil || claimed == nil || claimed.Status != entity.ScheduledMessageSending {
		t.Fatalf("予約中の予約を今すぐ送信で取り出せません: %+v %v", claimed, err)
	}
	if err := repo.MarkFailed(ctx, future.ID, "権限がありません"); err != nil {
		t.Fatalf("失敗を記録できません: %v", err)
	}
	if err := repo.Reschedule(ctx, future.ID, "直した", now.Add(2*time.Hour)); err != nil {
		t.Fatalf("日時を変えられません: %v", err)
	}
	rescheduled, _ := repo.FindByID(ctx, future.ID)
	if rescheduled.Status != entity.ScheduledMessageScheduled || rescheduled.FailureReason != nil || rescheduled.Body != "直した" {
		t.Errorf("失敗した予約が予約中に戻っていません: %+v", rescheduled)
	}
}
