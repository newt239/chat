package repository

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/newt239/chat/internal/domain/entity"
)

func TestAuditLogRepositoryListPaging(t *testing.T) {
	client := openTestClient(t)
	repo := NewAuditLogRepository(client)
	ctx := context.Background()
	workspaceID := "ws-" + uuid.NewString()[:8]

	base := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	// 同時刻の行があってもページの境目で取りこぼさないことを確かめる
	offsets := []time.Duration{0, time.Hour, time.Hour, time.Hour, 2 * time.Hour}
	for _, d := range offsets {
		log := &entity.AuditLog{WorkspaceID: workspaceID, Action: entity.AuditActionLogin, Metadata: map[string]string{}, CreatedAt: base.Add(d)}
		if err := repo.Create(ctx, log); err != nil {
			t.Fatalf("作成に失敗しました: %v", err)
		}
	}

	var got []*entity.AuditLog
	token := ""
	for range len(offsets) {
		page, err := repo.List(ctx, entity.AuditLogFilter{WorkspaceID: workspaceID, Limit: 2, PageToken: token})
		if err != nil {
			t.Fatalf("取得に失敗しました: %v", err)
		}
		got = append(got, page.Logs...)
		token = page.NextPageToken
		if token == "" {
			break
		}
	}

	if len(got) != len(offsets) {
		t.Fatalf("件数が一致しません: got=%d", len(got))
	}
	ids := make([]string, 0, len(got))
	for idx, l := range got {
		if slices.Contains(ids, l.ID) {
			t.Fatalf("同じ行が 2 回返されました: %s", l.ID)
		}
		ids = append(ids, l.ID)
		if idx > 0 && l.CreatedAt.After(got[idx-1].CreatedAt) {
			t.Fatalf("新しい順になっていません")
		}
	}

	if _, err := repo.List(ctx, entity.AuditLogFilter{WorkspaceID: workspaceID, Limit: 2, PageToken: "broken"}); !errors.Is(err, entity.ErrInvalidAuditLogPageToken) {
		t.Fatalf("壊れたトークンを受け付けました: %v", err)
	}
}
