package command

import (
	"errors"
	"testing"
	"time"
)

func TestParseRemind(t *testing.T) {
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Date(2026, 10, 1, 21, 30, 0, 0, tokyo)
	user := "11111111-1111-4111-8111-111111111111"
	channel := "22222222-2222-4222-8222-222222222222"

	tests := []struct {
		args        string
		wantText    string
		wantAt      time.Time
		wantUser    *string
		wantChannel *string
		wantErr     error
	}{
		{args: "me 資料を送る 10分後", wantText: "資料を送る", wantAt: now.Add(10 * time.Minute)},
		{args: "資料を送る 2時間後に", wantText: "資料を送る", wantAt: now.Add(2 * time.Hour)},
		{args: "report in 3d", wantText: "report", wantAt: now.Add(72 * time.Hour)},
		{args: "<@" + user + "> 「会議の準備」 明日 9:00", wantText: "会議の準備", wantAt: time.Date(2026, 10, 2, 9, 0, 0, 0, tokyo), wantUser: &user},
		{args: "<#" + channel + "> 締め切り 2026-10-05 18:00", wantText: "締め切り", wantAt: time.Date(2026, 10, 5, 18, 0, 0, 0, tokyo), wantChannel: &channel},
		{args: "年末 12/31 10:00", wantText: "年末", wantAt: time.Date(2026, 12, 31, 10, 0, 0, 0, tokyo)},
		{args: "始業 1/5 9:00", wantText: "始業", wantAt: time.Date(2027, 1, 5, 9, 0, 0, 0, tokyo)},
		{args: "寝る 23:00", wantText: "寝る", wantAt: time.Date(2026, 10, 1, 23, 0, 0, 0, tokyo)},
		{args: "起きる 7:00", wantText: "起きる", wantAt: time.Date(2026, 10, 2, 7, 0, 0, 0, tokyo)},
		{args: "日時がない", wantErr: ErrRemindUsage},
		{args: "10分後", wantErr: ErrRemindUsage},
		{args: "過去 今日 9:00", wantErr: ErrRemindInPast},
		{args: "存在しない日 2026/2/30 9:00", wantErr: ErrRemindUsage},
		{args: "遠すぎる 2028-01-01 9:00", wantErr: ErrRemindTooFar},
	}
	for _, tt := range tests {
		t.Run(tt.args, func(t *testing.T) {
			got, err := ParseRemind(tt.args, now)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got err=%v want=%v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Text != tt.wantText || !got.At.Equal(tt.wantAt) {
				t.Fatalf("got text=%q at=%s", got.Text, got.At)
			}
			if (got.Target.UserID == nil) != (tt.wantUser == nil) || (got.Target.ChannelID == nil) != (tt.wantChannel == nil) {
				t.Fatalf("宛先が期待と異なります: %+v", got.Target)
			}
		})
	}
}
