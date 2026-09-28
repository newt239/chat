package insight

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
	members map[string]*entity.WorkspaceMember
}

func (r *stubWorkspaceRepo) FindMember(_ context.Context, _ string, userID string) (*entity.WorkspaceMember, error) {
	return r.members[userID], nil
}

// 期間の開始日時で前期と今期を見分けて件数を返す
type stubInsightRepo struct {
	periodStart time.Time
	viewerID    *string
	dailyFrom   time.Time
	loc         *time.Location
}

func (r *stubInsightRepo) byPeriod(from time.Time, current, previous int) int {
	if from.Before(r.periodStart) {
		return previous
	}
	return current
}

func (r *stubInsightRepo) CountMessages(_ context.Context, _ string, from, _ time.Time) (int, error) {
	return r.byPeriod(from, 120, 100), nil
}

func (r *stubInsightRepo) CountActiveMembers(_ context.Context, _ string, from, _ time.Time) (int, error) {
	return r.byPeriod(from, 3, 2), nil
}

func (r *stubInsightRepo) CountMembersJoinedBefore(context.Context, string, time.Time) (int, error) {
	return 4, nil
}

func (r *stubInsightRepo) StorageBytesBefore(_ context.Context, _ string, before time.Time) (int64, error) {
	if before.Equal(r.periodStart) {
		return 1000, nil
	}
	return 1500, nil
}

func (r *stubInsightRepo) StorageByMimeType(context.Context, string) ([]entity.MimeTypeUsage, error) {
	return []entity.MimeTypeUsage{
		{MimeType: "image/png", Bytes: 100, FileCount: 2},
		{MimeType: "image/jpeg", Bytes: 50, FileCount: 1},
		{MimeType: "video/mp4", Bytes: 500, FileCount: 1},
		{MimeType: "application/pdf", Bytes: 30, FileCount: 3},
	}, nil
}

func (r *stubInsightRepo) DailyActivity(_ context.Context, _ string, from, _ time.Time, loc *time.Location) ([]entity.DailyActivity, error) {
	r.dailyFrom = from
	r.loc = loc
	return []entity.DailyActivity{{Date: "2026-09-28", MessageCount: 5, ActiveMemberCount: 2}}, nil
}

func (r *stubInsightRepo) DailyMessageCountsByUser(context.Context, string, string, time.Time, time.Time, *time.Location) ([]entity.DailyActivity, error) {
	return []entity.DailyActivity{{Date: "2026-09-27", MessageCount: 3}}, nil
}

func (r *stubInsightRepo) ChannelMessageCounts(_ context.Context, _ string, viewerID *string, _, _, _ time.Time) ([]entity.ChannelMessageCount, error) {
	r.viewerID = viewerID
	return nil, nil
}

func (r *stubInsightRepo) Heatmap(context.Context, string, time.Time, time.Time, *time.Location) ([]entity.HeatmapCell, error) {
	return []entity.HeatmapCell{{Weekday: 1, Hour: 9, MessageCount: 7}}, nil
}

func (r *stubInsightRepo) MemberActivities(context.Context, string, time.Time, time.Time) ([]entity.MemberActivity, error) {
	return nil, nil
}

func newInteractor(now time.Time) (*Interactor, *stubInsightRepo) {
	repo := &stubInsightRepo{periodStart: now.Add(-entity.InsightPeriod)}
	uc := NewInteractor(&stubWorkspaceRepo{members: map[string]*entity.WorkspaceMember{
		"admin":  {Role: entity.WorkspaceRoleAdmin},
		"member": {Role: entity.WorkspaceRoleMember},
	}}, repo)
	uc.now = func() time.Time { return now }
	return uc, repo
}

func TestGetInsights(t *testing.T) {
	now := time.Date(2026, 9, 28, 1, 0, 0, 0, time.UTC)
	uc, repo := newInteractor(now)

	out, err := uc.GetInsights(context.Background(), Input{WorkspaceID: "ws", RequesterID: "member", TimeZone: "Asia/Tokyo"})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	if out.MessageCount != (Comparison[int]{Current: 120, Previous: 100}) {
		t.Errorf("メッセージ数の比較が期待と異なります: %+v", out.MessageCount)
	}
	if out.ActiveRate != (Comparison[float64]{Current: 0.75, Previous: 0.5}) {
		t.Errorf("アクティブ率が期待と異なります: %+v", out.ActiveRate)
	}
	if out.StorageBytes != (Comparison[int64]{Current: 1500, Previous: 1000}) {
		t.Errorf("ストレージの比較が期待と異なります: %+v", out.StorageBytes)
	}

	// 東京では 9/28 10:00 なので、30 日分は 8/30 から 9/28 まで
	if repo.loc.String() != "Asia/Tokyo" || repo.dailyFrom.Format(time.DateTime) != "2026-08-30 00:00:00" {
		t.Errorf("日付の区切りが指定したタイムゾーンになっていません: %v %v", repo.loc, repo.dailyFrom)
	}
	if len(out.DailyActivity) != 30 || out.DailyActivity[0].Date != "2026-08-30" || out.DailyActivity[29] != (entity.DailyActivity{Date: "2026-09-28", MessageCount: 5, ActiveMemberCount: 2}) {
		t.Errorf("日別の集計が 0 件の日で埋められていません: %+v", out.DailyActivity)
	}
	if len(out.MyDailyMessages) != 7 || out.MyDailyMessages[5].MessageCount != 3 || out.MyDailyMessages[6].MessageCount != 0 {
		t.Errorf("自分の投稿数が期待と異なります: %+v", out.MyDailyMessages)
	}

	if len(out.Heatmap) != 7*24 || out.Heatmap[9] != (entity.HeatmapCell{Weekday: 1, Hour: 9, MessageCount: 7}) || out.Heatmap[10].MessageCount != 0 {
		t.Errorf("ヒートマップが全セルで埋められていません")
	}

	want := []StorageUsage{
		{Category: entity.StorageCategoryImage, Bytes: 150, FileCount: 3},
		{Category: entity.StorageCategoryVideo, Bytes: 500, FileCount: 1},
		{Category: entity.StorageCategoryAudio},
		{Category: entity.StorageCategoryFile, Bytes: 30, FileCount: 3},
	}
	for i, w := range want {
		if out.StorageBreakdown[i] != w {
			t.Errorf("ストレージの内訳が期待と異なります: got=%+v want=%+v", out.StorageBreakdown[i], w)
		}
	}

	if repo.viewerID == nil || *repo.viewerID != "member" {
		t.Errorf("メンバーには参加していない非公開チャンネルを見せてはいけません")
	}
}

func TestGetInsightsAdminSeesAllChannels(t *testing.T) {
	uc, repo := newInteractor(time.Now())
	if _, err := uc.GetInsights(context.Background(), Input{WorkspaceID: "ws", RequesterID: "admin", TimeZone: "UTC"}); err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if repo.viewerID != nil {
		t.Errorf("管理者はすべてのチャンネルを見られるはず")
	}
}

func TestGetInsightsRejects(t *testing.T) {
	uc, _ := newInteractor(time.Now())
	if _, err := uc.GetInsights(context.Background(), Input{WorkspaceID: "ws", RequesterID: "stranger", TimeZone: "UTC"}); !errors.Is(err, domerr.ErrUnauthorized) {
		t.Errorf("ワークスペース外のユーザーが閲覧できてしまいます: %v", err)
	}
	if _, err := uc.GetInsights(context.Background(), Input{WorkspaceID: "ws", RequesterID: "member", TimeZone: "Mars/Olympus"}); !errors.Is(err, ErrInvalidTimeZone) {
		t.Errorf("不正なタイムゾーンが拒否されていません: %v", err)
	}
}
