package entity

import "time"

type PushPlatform string

const (
	PushPlatformWeb     PushPlatform = "web"
	PushPlatformIOS     PushPlatform = "ios"
	PushPlatformAndroid PushPlatform = "android"
)

// PushToken はプッシュ通知を送る端末の FCM 登録トークンです
type PushToken struct {
	UserID     string
	Token      string
	Platform   PushPlatform
	UserAgent  string
	LastSeenAt time.Time
}
