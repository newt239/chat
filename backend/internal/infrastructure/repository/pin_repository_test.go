package repository

import (
	"context"
	"testing"
	"time"
)

func TestListPinsExcludesDeletedMessages(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	ctx := context.Background()
	dev := f.channels["dev"]
	deleted := client.Message.Create().SetChannel(dev).SetUser(f.bob).SetBody("消す").SetDeletedAt(time.Now()).SaveX(ctx)
	client.MessagePin.Create().SetChannel(dev).SetMessage(deleted).SetPinnedBy(f.bob).SaveX(ctx)

	pins, err := NewPinRepository(client).List(ctx, dev.ID.String(), 10)
	if err != nil {
		t.Fatalf("取得に失敗しました: %v", err)
	}
	if len(pins) != 1 || pins[0].MessageID != f.messages["image"].ID.String() {
		t.Errorf("削除したメッセージのピンが残っています: %+v", pins)
	}
}
