package presenter

import (
	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	insightuc "github.com/newt239/chat/internal/usecase/insight"
)

var storageCategories = map[entity.StorageCategory]chatv1.StorageCategory{
	entity.StorageCategoryImage: chatv1.StorageCategory_STORAGE_CATEGORY_IMAGE,
	entity.StorageCategoryVideo: chatv1.StorageCategory_STORAGE_CATEGORY_VIDEO,
	entity.StorageCategoryAudio: chatv1.StorageCategory_STORAGE_CATEGORY_AUDIO,
	entity.StorageCategoryFile:  chatv1.StorageCategory_STORAGE_CATEGORY_FILE,
}

func countComparison[T int | int64](c insightuc.Comparison[T]) *chatv1.CountComparison {
	return &chatv1.CountComparison{Current: int64(c.Current), Previous: int64(c.Previous)}
}

func Insights(out *insightuc.Output) *chatv1.GetInsightsResponse {
	return &chatv1.GetInsightsResponse{
		ActiveMembers: countComparison(out.ActiveMembers),
		MemberCount:   countComparison(out.MemberCount),
		ActiveRate:    &chatv1.RatioComparison{Current: out.ActiveRate.Current, Previous: out.ActiveRate.Previous},
		MessageCount:  countComparison(out.MessageCount),
		StorageBytes:  countComparison(out.StorageBytes),
		DailyActivity: ConvertAll(out.DailyActivity, func(d entity.DailyActivity) *chatv1.DailyActivity {
			return &chatv1.DailyActivity{Date: d.Date, MessageCount: int32(d.MessageCount), ActiveMemberCount: int32(d.ActiveMemberCount)}
		}),
		Channels: ConvertAll(out.Channels, func(c entity.ChannelMessageCount) *chatv1.ChannelActivity {
			return &chatv1.ChannelActivity{
				ChannelId:          c.ChannelID,
				Name:               c.Name,
				IsPrivate:          c.IsPrivate,
				MessageCount:       int32(c.MessageCount),
				RecentMessageCount: int32(c.RecentMessageCount),
			}
		}),
		Heatmap: ConvertAll(out.Heatmap, func(c entity.HeatmapCell) *chatv1.HeatmapCell {
			return &chatv1.HeatmapCell{Weekday: int32(c.Weekday), Hour: int32(c.Hour), MessageCount: int32(c.MessageCount)}
		}),
		HeatmapWeeks: insightuc.HeatmapWeeks,
		StorageBreakdown: ConvertAll(out.StorageBreakdown, func(u insightuc.StorageUsage) *chatv1.StorageUsage {
			return &chatv1.StorageUsage{Category: storageCategories[u.Category], Bytes: u.Bytes, FileCount: int32(u.FileCount)}
		}),
		MyDailyMessages: ConvertAll(out.MyDailyMessages, func(d entity.DailyActivity) *chatv1.DailyCount {
			return &chatv1.DailyCount{Date: d.Date, Count: int32(d.MessageCount)}
		}),
	}
}
