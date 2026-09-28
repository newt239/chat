// Package insight はメンバー全員が閲覧できるワークスペースの指標を集計するユースケースです
package insight

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

var ErrInvalidTimeZone = errors.New("タイムゾーンの指定が正しくありません")

const (
	dailyActivityDays = 30
	myDailyDays       = 7
	recentDays        = 7
	// HeatmapWeeks はヒートマップを集計する週数です
	HeatmapWeeks = 4
)

type Input struct {
	WorkspaceID string
	RequesterID string
	// IANA のタイムゾーン名。日付と曜日・時刻の区切りに使う
	TimeZone string
}

type Comparison[T int | int64 | float64] struct {
	Current  T
	Previous T
}

type StorageUsage struct {
	Category  entity.StorageCategory
	Bytes     int64
	FileCount int
}

type Output struct {
	ActiveMembers    Comparison[int]
	MemberCount      Comparison[int]
	ActiveRate       Comparison[float64]
	MessageCount     Comparison[int]
	StorageBytes     Comparison[int64]
	DailyActivity    []entity.DailyActivity
	Channels         []entity.ChannelMessageCount
	Heatmap          []entity.HeatmapCell
	StorageBreakdown []StorageUsage
	MyDailyMessages  []entity.DailyActivity
}

type Interactor struct {
	workspaceRepo domainrepository.WorkspaceRepository
	insightRepo   domainrepository.InsightRepository
	now           func() time.Time
}

func NewInteractor(workspaceRepo domainrepository.WorkspaceRepository, insightRepo domainrepository.InsightRepository) *Interactor {
	return &Interactor{workspaceRepo: workspaceRepo, insightRepo: insightRepo, now: time.Now}
}

// GetInsights は個人を特定できる数値を含めずに指標を返します（本人の投稿数だけは例外）
func (i *Interactor) GetInsights(ctx context.Context, input Input) (*Output, error) {
	loc, err := time.LoadLocation(input.TimeZone)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidTimeZone, input.TimeZone)
	}
	member, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, domerr.ErrUnauthorized
	}

	now := i.now().In(loc)
	periodStart := now.Add(-entity.InsightPeriod)
	previousStart := periodStart.Add(-entity.InsightPeriod)
	ws := input.WorkspaceID
	out := &Output{}

	if out.ActiveMembers, err = compare(ctx, func(ctx context.Context, from, to time.Time) (int, error) {
		return i.insightRepo.CountActiveMembers(ctx, ws, from, to)
	}, previousStart, periodStart, now); err != nil {
		return nil, err
	}
	if out.MessageCount, err = compare(ctx, func(ctx context.Context, from, to time.Time) (int, error) {
		return i.insightRepo.CountMessages(ctx, ws, from, to)
	}, previousStart, periodStart, now); err != nil {
		return nil, err
	}
	if out.MemberCount, err = compare(ctx, func(ctx context.Context, _, to time.Time) (int, error) {
		return i.insightRepo.CountMembersJoinedBefore(ctx, ws, to)
	}, previousStart, periodStart, now); err != nil {
		return nil, err
	}
	if out.StorageBytes, err = compare(ctx, func(ctx context.Context, _, to time.Time) (int64, error) {
		return i.insightRepo.StorageBytesBefore(ctx, ws, to)
	}, previousStart, periodStart, now); err != nil {
		return nil, err
	}
	out.ActiveRate = Comparison[float64]{
		Current:  rate(out.ActiveMembers.Current, out.MemberCount.Current),
		Previous: rate(out.ActiveMembers.Previous, out.MemberCount.Previous),
	}

	dailyStart := startOfDay(now).AddDate(0, 0, -(dailyActivityDays - 1))
	daily, err := i.insightRepo.DailyActivity(ctx, ws, dailyStart, now, loc)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate daily activity: %w", err)
	}
	out.DailyActivity = fillDays(daily, dailyStart, dailyActivityDays)

	myStart := startOfDay(now).AddDate(0, 0, -(myDailyDays - 1))
	mine, err := i.insightRepo.DailyMessageCountsByUser(ctx, ws, input.RequesterID, myStart, now, loc)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate own messages: %w", err)
	}
	out.MyDailyMessages = fillDays(mine, myStart, myDailyDays)

	// 管理者は非公開チャンネルもすべて見られる。それ以外は参加している非公開チャンネルだけ
	var viewerID *string
	if !member.IsAdmin() {
		viewerID = &input.RequesterID
	}
	if out.Channels, err = i.insightRepo.ChannelMessageCounts(ctx, ws, viewerID, periodStart, now.AddDate(0, 0, -recentDays), now); err != nil {
		return nil, fmt.Errorf("failed to aggregate channel messages: %w", err)
	}

	cells, err := i.insightRepo.Heatmap(ctx, ws, now.AddDate(0, 0, -7*HeatmapWeeks), now, loc)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate heatmap: %w", err)
	}
	out.Heatmap = fillHeatmap(cells)

	usages, err := i.insightRepo.StorageByMimeType(ctx, ws)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate storage: %w", err)
	}
	out.StorageBreakdown = storageBreakdown(usages)

	return out, nil
}

func compare[T int | int64](ctx context.Context, count func(ctx context.Context, from, to time.Time) (T, error), previousStart, periodStart, now time.Time) (Comparison[T], error) {
	current, err := count(ctx, periodStart, now)
	if err != nil {
		return Comparison[T]{}, fmt.Errorf("failed to aggregate current period: %w", err)
	}
	previous, err := count(ctx, previousStart, periodStart)
	if err != nil {
		return Comparison[T]{}, fmt.Errorf("failed to aggregate previous period: %w", err)
	}
	return Comparison[T]{Current: current, Previous: previous}, nil
}

func rate(active, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(active) / float64(total)
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// fillDays は集計結果に含まれない日を 0 件で埋めます
func fillDays(rows []entity.DailyActivity, start time.Time, days int) []entity.DailyActivity {
	byDate := make(map[string]entity.DailyActivity, len(rows))
	for _, r := range rows {
		byDate[r.Date] = r
	}
	result := make([]entity.DailyActivity, 0, days)
	for d := range days {
		date := start.AddDate(0, 0, d).Format(time.DateOnly)
		row := byDate[date]
		row.Date = date
		result = append(result, row)
	}
	return result
}

func fillHeatmap(cells []entity.HeatmapCell) []entity.HeatmapCell {
	counts := make(map[[2]int]int, len(cells))
	for _, c := range cells {
		counts[[2]int{c.Weekday, c.Hour}] = c.MessageCount
	}
	result := make([]entity.HeatmapCell, 0, 7*24)
	for weekday := 1; weekday <= 7; weekday++ {
		for hour := range 24 {
			result = append(result, entity.HeatmapCell{Weekday: weekday, Hour: hour, MessageCount: counts[[2]int{weekday, hour}]})
		}
	}
	return result
}

func storageBreakdown(usages []entity.MimeTypeUsage) []StorageUsage {
	categories := []entity.StorageCategory{entity.StorageCategoryImage, entity.StorageCategoryVideo, entity.StorageCategoryAudio, entity.StorageCategoryFile}
	totals := make(map[entity.StorageCategory]StorageUsage, len(categories))
	for _, u := range usages {
		category := entity.StorageCategoryOf(u.MimeType)
		total := totals[category]
		total.Bytes += u.Bytes
		total.FileCount += u.FileCount
		totals[category] = total
	}
	result := make([]StorageUsage, 0, len(categories))
	for _, c := range categories {
		usage := totals[c]
		usage.Category = c
		result = append(result, usage)
	}
	return result
}
