package entity

import (
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// MessagePermalink はフロントエンドのメッセージ URL が指すメッセージです
type MessagePermalink struct {
	WorkspaceID string
	ChannelID   string
	MessageID   string
}

// ParseMessagePermalink は /app/{ws}/{ch}?message={id} と /app/{ws}/{ch}/thread/{threadId}[?message={replyId}] を解釈します
func ParseMessagePermalink(rawURL string) (MessagePermalink, bool) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return MessagePermalink{}, false
	}
	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(segments) < 3 || segments[0] != "app" || segments[1] == "" || !isUUID(segments[2]) {
		return MessagePermalink{}, false
	}

	messageID := u.Query().Get("message")
	switch {
	case len(segments) == 5 && segments[3] == "thread" && isUUID(segments[4]):
		if messageID == "" {
			messageID = segments[4]
		}
	case len(segments) != 3:
		return MessagePermalink{}, false
	}
	if !isUUID(messageID) {
		return MessagePermalink{}, false
	}

	return MessagePermalink{WorkspaceID: segments[1], ChannelID: segments[2], MessageID: messageID}, true
}

// MessagePermalinkPath はメッセージを開くフロントエンドのパスを返します。スレッドの返信はスレッドの中で開きます
func MessagePermalinkPath(workspaceID, channelID, messageID string, parentID *string) string {
	base := "/app/" + url.PathEscape(workspaceID) + "/" + channelID
	if parentID != nil {
		base += "/thread/" + *parentID
	}
	return base + "?message=" + messageID
}

func isUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
