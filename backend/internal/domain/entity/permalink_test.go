package entity

import "testing"

func TestParseMessagePermalink(t *testing.T) {
	const (
		channelID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
		messageID = "f1111111-1111-1111-1111-111111111111"
		replyID   = "f2222222-2222-2222-2222-222222222222"
	)

	tests := []struct {
		name   string
		url    string
		want   MessagePermalink
		wantOK bool
	}{
		{
			name:   "チャンネルのメッセージ",
			url:    "https://chat.example.com/app/general/" + channelID + "?message=" + messageID,
			want:   MessagePermalink{WorkspaceID: "general", ChannelID: channelID, MessageID: messageID},
			wantOK: true,
		},
		{
			name:   "スレッドは親メッセージを指す",
			url:    "http://localhost:5173/app/general/" + channelID + "/thread/" + messageID,
			want:   MessagePermalink{WorkspaceID: "general", ChannelID: channelID, MessageID: messageID},
			wantOK: true,
		},
		{
			name:   "スレッド内の返信",
			url:    "http://localhost:5173/app/general/" + channelID + "/thread/" + messageID + "?message=" + replyID,
			want:   MessagePermalink{WorkspaceID: "general", ChannelID: channelID, MessageID: replyID},
			wantOK: true,
		},
		{name: "メッセージ ID がない", url: "https://chat.example.com/app/general/" + channelID},
		{name: "メッセージ ID が UUID でない", url: "https://chat.example.com/app/general/" + channelID + "?message=abc"},
		{name: "チャンネル ID が UUID でない", url: "https://chat.example.com/app/general/random?message=" + messageID},
		{name: "アプリ以外のパス", url: "https://example.com/blog/" + channelID + "?message=" + messageID},
		{name: "余分なパス", url: "https://chat.example.com/app/general/" + channelID + "/files?message=" + messageID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseMessagePermalink(tt.url)
			if ok != tt.wantOK || got != tt.want {
				t.Errorf("got=(%+v, %t) want=(%+v, %t)", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
