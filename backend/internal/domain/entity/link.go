package entity

import "time"

type MessageLink struct {
	ID        string
	MessageID string
	URL       string
	OGP       OGPData
	// 同じワークスペースのメッセージへのリンクのときに設定される
	LinkedMessageID *string
	CreatedAt       time.Time
}

type OGPData struct {
	Title       *string
	Description *string
	ImageURL    *string
	SiteName    *string
	CardType    *string
	ImageWidth  *int32
	ImageHeight *int32
	YouTube     *YouTubeVideo
}

type YouTubeVideo struct {
	VideoID         string
	ChannelName     *string
	DurationSeconds *int32
}
