package notification

import (
	"log"
	"reflect"

	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/interfaces/handler/websocket"
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

// NotifyNewMessage は新しいメッセージをチャンネル購読者に通知します
func (s *WebSocketNotificationService) NotifyNewMessage(workspaceID string, channelID string, message interface{}) {
	payload := websocket.NewMessagePayload{
		ChannelID: channelID,
		Message:   convertToMap(message),
	}

	data, err := websocket.SendServerMessage(websocket.EventTypeNewMessage, payload)
	if err != nil {
		log.Printf("new_messageイベントのエンコードに失敗しました: %v", err)
		return
	}

	s.hub.BroadcastToChannelSubscribers(workspaceID, channelID, data)
	log.Printf("Notified new message to workspace=%s channel=%s", workspaceID, channelID)
}

// NotifyReaction はリアクション追加をチャンネル購読者に通知します
func (s *WebSocketNotificationService) NotifyReaction(workspaceID string, channelID string, reaction interface{}) {
	// リアクションは new_message イベントの一種として扱う
	// 将来的に専用のイベントタイプを追加することも検討
	payload := websocket.NewMessagePayload{
		ChannelID: channelID,
		Message:   convertToMap(reaction),
	}

	data, err := websocket.SendServerMessage(websocket.EventTypeNewMessage, payload)
	if err != nil {
		log.Printf("reactionイベントのエンコードに失敗しました: %v", err)
		return
	}

	s.hub.BroadcastToChannelSubscribers(workspaceID, channelID, data)
	log.Printf("Notified reaction to workspace=%s channel=%s", workspaceID, channelID)
}

// NotifyUpdatedMessage はメッセージ更新をチャンネル購読者に通知します
func (s *WebSocketNotificationService) NotifyUpdatedMessage(workspaceID string, channelID string, message interface{}) {
	payload := websocket.MessageUpdatedPayload{
		ChannelID: channelID,
		Message:   convertToMap(message),
	}

	data, err := websocket.SendServerMessage(websocket.EventTypeMessageUpdated, payload)
	if err != nil {
		log.Printf("message_updatedイベントのエンコードに失敗しました: %v", err)
		return
	}

	s.hub.BroadcastToChannelSubscribers(workspaceID, channelID, data)
	log.Printf("Notified message updated to workspace=%s channel=%s", workspaceID, channelID)
}

// NotifyDeletedMessage はメッセージ削除をチャンネル購読者に通知します
func (s *WebSocketNotificationService) NotifyDeletedMessage(workspaceID string, channelID string, deleteData interface{}) {
	payload := websocket.MessageDeletedPayload{
		ChannelID:  channelID,
		DeleteData: convertToMap(deleteData),
	}

	data, err := websocket.SendServerMessage(websocket.EventTypeMessageDeleted, payload)
	if err != nil {
		log.Printf("message_deletedイベントのエンコードに失敗しました: %v", err)
		return
	}

	s.hub.BroadcastToChannelSubscribers(workspaceID, channelID, data)
	log.Printf("Notified message deleted to workspace=%s channel=%s", workspaceID, channelID)
}

// NotifyPinCreated はピン追加をチャンネル参加者に通知します
func (s *WebSocketNotificationService) NotifyPinCreated(workspaceID string, channelID string, pin interface{}) {
	payload := websocket.PinPayload{
		ChannelID: channelID,
		Message:   convertToMap(pin),
		// PinnedBy/PinnedAt は pin の中に含めている前提。トップレベルにも複製しておく
	}
	if m, ok := pin.(map[string]interface{}); ok {
		if v, ok := m["pinnedBy"].(string); ok {
			payload.PinnedBy = v
		}
		if v, ok := m["pinnedAt"].(string); ok {
			payload.PinnedAt = v
		}
	}
	data, err := websocket.SendServerMessage(websocket.EventTypePinCreated, payload)
	if err != nil {
		log.Printf("pin_createdイベントのエンコードに失敗しました: %v", err)
		return
	}
	s.hub.BroadcastToChannel(workspaceID, channelID, data)
}

// NotifyPinDeleted はピン削除をチャンネル参加者に通知します
func (s *WebSocketNotificationService) NotifyPinDeleted(workspaceID string, channelID string, pin interface{}) {
	payload := websocket.PinPayload{
		ChannelID: channelID,
		Message:   convertToMap(pin),
	}
	if m, ok := pin.(map[string]interface{}); ok {
		if v, ok := m["pinnedBy"].(string); ok {
			payload.PinnedBy = v
		}
		if v, ok := m["pinnedAt"].(string); ok {
			payload.PinnedAt = v
		}
	}
	data, err := websocket.SendServerMessage(websocket.EventTypePinDeleted, payload)
	if err != nil {
		log.Printf("pin_deletedイベントのエンコードに失敗しました: %v", err)
		return
	}
	s.hub.BroadcastToChannel(workspaceID, channelID, data)
}

// NotifyUnreadCount は未読数の更新を特定ユーザーに通知します
func (s *WebSocketNotificationService) NotifyUnreadCount(workspaceID string, userID string, channelID string, unreadCount int) {
	// TODO: メンション検知の実装（現在は未読数が0より大きい場合にtrueとする）
	hasMention := unreadCount > 0

	payload := websocket.UnreadCountPayload{
		ChannelID:   channelID,
		UnreadCount: unreadCount,
		HasMention:  hasMention,
	}

	data, err := websocket.SendServerMessage(websocket.EventTypeUnreadCount, payload)
	if err != nil {
		log.Printf("unread_countイベントのエンコードに失敗しました: %v", err)
		return
	}

	s.hub.BroadcastToUser(workspaceID, userID, data)
	log.Printf("Notified unread count to workspace=%s user=%s channel=%s count=%d mention=%t", workspaceID, userID, channelID, unreadCount, hasMention)
}

// convertToMap は任意の構造体をmap[string]interface{}に変換します
func convertToMap(data interface{}) map[string]interface{} {
	// データが既にマップの場合はそのまま返す
	if m, ok := data.(map[string]interface{}); ok {
		return m
	}

	// リフレクションを使った効率的な変換
	return convertStructToMap(data)
}

// convertStructToMap はリフレクションを使って構造体をmap[string]interface{}に変換します
func convertStructToMap(data interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		log.Printf("警告: データが構造体ではありません (kind=%v)", v.Kind())
		return result
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// フィールド名を取得（jsonタグがあればそれを使用）
		fieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			// jsonタグからカンマ前の部分を取得
			if commaIndex := len(jsonTag); commaIndex > 0 {
				for j, c := range jsonTag {
					if c == ',' {
						commaIndex = j
						break
					}
				}
				fieldName = jsonTag[:commaIndex]
			}
		}

		// フィールドが公開されている場合のみ処理
		if fieldValue.CanInterface() {
			result[fieldName] = fieldValue.Interface()
		}
	}

	return result
}
