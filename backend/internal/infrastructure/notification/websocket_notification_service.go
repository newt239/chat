package notification

import (
	"encoding/json"
	"log"

	"github.com/example/chat/internal/adapter/controller/websocket"
	"github.com/example/chat/internal/domain/service"
)

// WebSocketNotificationService はWebSocketを利用した通知サービスの実装です
type WebSocketNotificationService struct {
	hub *websocket.Hub
}

// NewWebSocketNotificationService は新しいWebSocketNotificationServiceを作成します
func NewWebSocketNotificationService(hub *websocket.Hub) service.NotificationService {
	return &WebSocketNotificationService{
		hub: hub,
	}
}

// NotifyNewMessage は新しいメッセージをチャンネル参加者に通知します
func (s *WebSocketNotificationService) NotifyNewMessage(workspaceID string, channelID string, message interface{}) {
	payload := websocket.NewMessagePayload{
		ChannelID: channelID,
		Message:   convertToMap(message),
	}

	data, err := websocket.SendServerMessage(websocket.EventTypeNewMessage, payload)
	if err != nil {
		log.Printf("Failed to encode new_message event: %v", err)
		return
	}

	s.hub.BroadcastToChannel(workspaceID, channelID, data)
	log.Printf("Notified new message to workspace=%s channel=%s", workspaceID, channelID)
}

// NotifyReaction はリアクション追加をチャンネル参加者に通知します
func (s *WebSocketNotificationService) NotifyReaction(workspaceID string, channelID string, reaction interface{}) {
	// リアクションは new_message イベントの一種として扱う
	// 将来的に専用のイベントタイプを追加することも検討
	payload := websocket.NewMessagePayload{
		ChannelID: channelID,
		Message:   convertToMap(reaction),
	}

	data, err := websocket.SendServerMessage(websocket.EventTypeNewMessage, payload)
	if err != nil {
		log.Printf("Failed to encode reaction event: %v", err)
		return
	}

	s.hub.BroadcastToChannel(workspaceID, channelID, data)
	log.Printf("Notified reaction to workspace=%s channel=%s", workspaceID, channelID)
}

// NotifyUpdatedMessage はメッセージ更新をチャンネル参加者に通知します
func (s *WebSocketNotificationService) NotifyUpdatedMessage(workspaceID string, channelID string, message interface{}) {
	payload := websocket.MessageUpdatedPayload{
		ChannelID: channelID,
		Message:   convertToMap(message),
	}

	data, err := websocket.SendServerMessage(websocket.EventTypeMessageUpdated, payload)
	if err != nil {
		log.Printf("Failed to encode message_updated event: %v", err)
		return
	}

	s.hub.BroadcastToChannel(workspaceID, channelID, data)
	log.Printf("Notified message updated to workspace=%s channel=%s", workspaceID, channelID)
}

// NotifyDeletedMessage はメッセージ削除をチャンネル参加者に通知します
func (s *WebSocketNotificationService) NotifyDeletedMessage(workspaceID string, channelID string, deleteData interface{}) {
	payload := websocket.MessageDeletedPayload{
		ChannelID:  channelID,
		DeleteData: convertToMap(deleteData),
	}

	data, err := websocket.SendServerMessage(websocket.EventTypeMessageDeleted, payload)
	if err != nil {
		log.Printf("Failed to encode message_deleted event: %v", err)
		return
	}

	s.hub.BroadcastToChannel(workspaceID, channelID, data)
	log.Printf("Notified message deleted to workspace=%s channel=%s", workspaceID, channelID)
}

// NotifyUnreadCount は未読数の更新を特定ユーザーに通知します
func (s *WebSocketNotificationService) NotifyUnreadCount(workspaceID string, userID string, channelID string, unreadCount int) {
	payload := websocket.UnreadCountPayload{
		ChannelID:   channelID,
		UnreadCount: unreadCount,
	}

	data, err := websocket.SendServerMessage(websocket.EventTypeUnreadCount, payload)
	if err != nil {
		log.Printf("Failed to encode unread_count event: %v", err)
		return
	}

	s.hub.BroadcastToUser(workspaceID, userID, data)
	log.Printf("Notified unread count to workspace=%s user=%s channel=%s count=%d", workspaceID, userID, channelID, unreadCount)
}

// convertToMap は任意の構造体をmap[string]interface{}に変換します
func convertToMap(data interface{}) map[string]interface{} {
	// データが既にマップの場合はそのまま返す
	if m, ok := data.(map[string]interface{}); ok {
		return m
	}

	// JSON経由で変換
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Warning: failed to marshal data to JSON: %v", err)
		return map[string]interface{}{}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		log.Printf("Warning: failed to unmarshal JSON to map: %v", err)
		return map[string]interface{}{}
	}

	return result
}
