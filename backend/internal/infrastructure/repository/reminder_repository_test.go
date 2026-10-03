package repository

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent/reminder"
	"github.com/newt239/chat/internal/domain/entity"
)

func TestReminderClaimDueFailsStaleSending(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	repo := NewReminderRepository(client)
	ctx := context.Background()
	now := time.Now()

	due := &entity.Reminder{WorkspaceID: f.workspaceID, CreatorID: f.alice.ID.String(), Text: "remind", RemindAt: now.Add(-time.Minute)}
	if err := repo.Create(ctx, due); err != nil {
		t.Fatalf("リマインダーを作れません: %v", err)
	}
	claimed, err := repo.ClaimDue(ctx, now, now.Add(-time.Hour), 100)
	if err != nil || !slices.ContainsFunc(claimed, func(r *entity.Reminder) bool { return r.ID == due.ID }) {
		t.Fatalf("期限の来たリマインダーを取り出せません: %v", err)
	}

	// 送信中のまま止まったものは失敗にして、二重に届けない
	if _, err := repo.ClaimDue(ctx, now, time.Now().Add(time.Minute), 100); err != nil {
		t.Fatalf("取り出しに失敗しました: %v", err)
	}
	if got := client.Reminder.GetX(ctx, uuid.MustParse(due.ID)); got.Status != reminder.StatusFailed {
		t.Errorf("送信中のまま止まったリマインダーが失敗になっていません: %s", got.Status)
	}
}
