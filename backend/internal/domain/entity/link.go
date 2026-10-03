package entity

import (
	"regexp"
	"slices"
	"time"
)

type MessageLink struct {
	ID        string
	MessageID string
	URL       string
	OGP       OGPData
	// 外部の URL のときに設定される、URL ごとに共有するプレビュー
	LinkPreviewID *string
	// 同じワークスペースのメッセージへのリンクのときに設定される
	LinkedMessageID *string
	CreatedAt       time.Time
}

// LinkPreviewTTL を過ぎたプレビューは次に投稿されたときに取り直す
const LinkPreviewTTL = 24 * time.Hour

// LinkPreview は URL ごとに 1 つ保存する OGP です。取得に失敗した URL も空の OGP で保存し、取り直しを抑える
type LinkPreview struct {
	ID        string
	URL       string
	OGP       OGPData
	FetchedAt time.Time
}

// IsStale はプレビューを取り直す時期かを返します
func (p *LinkPreview) IsStale(now time.Time) bool {
	return now.Sub(p.FetchedAt) > LinkPreviewTTL
}

type OGPData struct {
	Title       *string
	Description *string
	ImageURL    *string
	SiteName    *string
	CardType    *string
	YouTube     *YouTubeVideo
	XPost       *XPost
}

type XPost struct {
	AuthorName   string
	AuthorHandle string
}

type YouTubeVideo struct {
	VideoID         string
	ChannelName     *string
	DurationSeconds *int32
}

var urlPattern = regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`)

// ExtractURLs は本文中の URL を出現順に重複なく返します
func ExtractURLs(text string) []string {
	var urls []string
	for _, match := range urlPattern.FindAllString(text, -1) {
		if !slices.Contains(urls, match) {
			urls = append(urls, match)
		}
	}
	return urls
}
