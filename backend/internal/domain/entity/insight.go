package entity

import (
	"strings"
	"time"
)

// InsightPeriod は KPI を集計する期間で、前期との比較にも同じ長さを使います
const InsightPeriod = 30 * 24 * time.Hour

type DailyActivity struct {
	// タイムゾーンで区切った日付 (YYYY-MM-DD)
	Date              string
	MessageCount      int
	ActiveMemberCount int
}

type ChannelMessageCount struct {
	ChannelID          string
	Name               string
	IsPrivate          bool
	MessageCount       int
	RecentMessageCount int
}

type HeatmapCell struct {
	// ISO 8601 の曜日 (1 = 月曜日, 7 = 日曜日)
	Weekday      int
	Hour         int
	MessageCount int
}

type MimeTypeUsage struct {
	MimeType  string
	Bytes     int64
	FileCount int
}

type StorageCategory string

const (
	StorageCategoryImage StorageCategory = "image"
	StorageCategoryVideo StorageCategory = "video"
	StorageCategoryAudio StorageCategory = "audio"
	StorageCategoryFile  StorageCategory = "file"
)

func StorageCategoryOf(mimeType string) StorageCategory {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return StorageCategoryImage
	case strings.HasPrefix(mimeType, "video/"):
		return StorageCategoryVideo
	case strings.HasPrefix(mimeType, "audio/"):
		return StorageCategoryAudio
	default:
		return StorageCategoryFile
	}
}

type MemberActivity struct {
	UserID        string
	MessageCount  int
	StorageBytes  int64
	LastMessageAt *time.Time
}
