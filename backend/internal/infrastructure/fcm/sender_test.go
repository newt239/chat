package fcm

import (
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	notificationuc "github.com/newt239/chat/internal/usecase/notification"
)

func TestToFCMMessage(t *testing.T) {
	base := notificationuc.PushMessage{Token: "t", Title: "Alice", Body: "hi", Data: map[string]string{"channelId": "c1"}}

	web := base
	web.Platform = entity.PushPlatformWeb
	got := toFCMMessage(web)
	if got.Data["title"] != "Alice" || got.Data["body"] != "hi" || got.Notification != nil || got.Webpush != nil {
		t.Errorf("ウェブはデータだけで送る: %+v", got)
	}
	if _, ok := base.Data["title"]; ok {
		t.Error("元のデータを書き換えています")
	}

	android := base
	android.Platform = entity.PushPlatformAndroid
	if got := toFCMMessage(android); got.Android == nil || got.Android.Notification.Title != "Alice" {
		t.Errorf("Android に通知の表示内容がありません: %+v", got)
	}

	ios := base
	ios.Platform = entity.PushPlatformIOS
	if got := toFCMMessage(ios); got.APNS == nil || got.APNS.Payload.Aps.Alert.Body != "hi" {
		t.Errorf("iOS に通知の表示内容がありません: %+v", got)
	}
}
