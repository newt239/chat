package websocket

import (
	"encoding/json"
	"fmt"
)

// EventType はWebSocketイベントのタイプを表します
type EventType string

const (
	// クライアント→サーバー
	EventTypeJoinChannel     EventType = "join_channel"
	EventTypeLeaveChannel    EventType = "leave_channel"
	EventTypePostMessage     EventType = "post_message"
	EventTypeTyping          EventType = "typing"
	EventTypeUpdateReadState EventType = "update_read_state"

	// サーバー→クライアント
	EventTypeNewMessage     EventType = "new_message"
	EventTypeMessageUpdated EventType = "message_updated"
	EventTypeMessageDeleted EventType = "message_deleted"
	EventTypeUnreadCount    EventType = "unread_count"
	EventTypeAck            EventType = "ack"
	EventTypeError          EventType = "error"
)

// ClientMessage はクライアントから受信するメッセージを表します
type ClientMessage struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// ServerMessage はサーバーからクライアントに送信するメッセージを表します
type ServerMessage struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

// JoinChannelPayload はjoin_channelイベントのペイロードを表します
type JoinChannelPayload struct {
	ChannelID string `json:"channel_id"`
}

// LeaveChannelPayload はleave_channelイベントのペイロードを表します
type LeaveChannelPayload struct {
	ChannelID string `json:"channel_id"`
}

// PostMessagePayload はpost_messageイベントのペイロードを表します
type PostMessagePayload struct {
	ChannelID string `json:"channel_id"`
	Body      string `json:"body"`
}

// TypingPayload はtypingイベントのペイロードを表します
type TypingPayload struct {
	ChannelID string `json:"channel_id"`
}

// UpdateReadStatePayload はupdate_read_stateイベントのペイロードを表します
type UpdateReadStatePayload struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
}

// NewMessagePayload はnew_messageイベントのペイロードを表します
type NewMessagePayload struct {
	ChannelID string                 `json:"channel_id"`
	Message   map[string]interface{} `json:"message"`
}

// MessageUpdatedPayload はmessage_updatedイベントのペイロードを表します
type MessageUpdatedPayload struct {
	ChannelID string                 `json:"channel_id"`
	Message   map[string]interface{} `json:"message"`
}

// MessageDeletedPayload はmessage_deletedイベントのペイロードを表します
type MessageDeletedPayload struct {
	ChannelID  string                 `json:"channel_id"`
	DeleteData map[string]interface{} `json:"deleteData"`
}

// UnreadCountPayload はunread_countイベントのペイロードを表します
type UnreadCountPayload struct {
	ChannelID   string `json:"channel_id"`
	UnreadCount int    `json:"unread_count"`
}

// AckPayload はackイベントのペイロードを表します
type AckPayload struct {
	Type    EventType `json:"type"`
	Success bool      `json:"success"`
	Message string    `json:"message,omitempty"`
}

// ErrorPayload はerrorイベントのペイロードを表します
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SendServerMessage はサーバーメッセージをJSON形式にエンコードします
func SendServerMessage(eventType EventType, payload interface{}) ([]byte, error) {
	msg := ServerMessage{
		Type:    eventType,
		Payload: payload,
	}
	return json.Marshal(msg)
}

// ParseClientMessage はクライアントメッセージをパースします
func ParseClientMessage(data []byte) (*ClientMessage, error) {
	var msg ClientMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}
	return &msg, nil
}
