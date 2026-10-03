package entity

type PushPlatform string

const (
	PushPlatformWeb     PushPlatform = "web"
	PushPlatformIOS     PushPlatform = "ios"
	PushPlatformAndroid PushPlatform = "android"
)

// PushToken はプッシュ通知を送る端末の FCM の送信先 ID（Firebase Installation ID）です
type PushToken struct {
	UserID   string
	Token    string
	Platform PushPlatform
}
