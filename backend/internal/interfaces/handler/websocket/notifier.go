package websocket

import (
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/infrastructure/logger"
	"github.com/newt239/chat/internal/interfaces/presenter"
	customemojiuc "github.com/newt239/chat/internal/usecase/customemoji"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	pinuc "github.com/newt239/chat/internal/usecase/pin"
	reactionuc "github.com/newt239/chat/internal/usecase/reaction"
)

// encodeServerEvent はイベントを JSON に変換します。失敗した場合は nil を返します
func encodeServerEvent(event *chatv1.ServerEvent) []byte {
	data, err := protojson.Marshal(event)
	if err != nil {
		logger.Get().Error("イベントのエンコードに失敗しました", zap.Error(err))
		return nil
	}
	return data
}

// Notifier は各ユースケースの変更通知を WebSocket のイベントとして配信します
type Notifier struct {
	hub *Hub
}

func NewNotifier(hub *Hub) *Notifier {
	return &Notifier{hub: hub}
}

func (n *Notifier) broadcastToSubscribers(workspaceID, channelID string, event *chatv1.ServerEvent) {
	if data := encodeServerEvent(event); data != nil {
		n.hub.BroadcastToChannel(workspaceID, channelID, data, "")
	}
}

func (n *Notifier) broadcastToUsers(workspaceID string, userIDs []string, event *chatv1.ServerEvent) {
	if data := encodeServerEvent(event); data != nil {
		n.hub.BroadcastToUsers(workspaceID, userIDs, data)
	}
}

// RevokeChannel はチャンネルを見られなくなった接続の購読を外します。userID が空なら購読者全員を確かめ直します
func (n *Notifier) RevokeChannel(workspaceID, channelID, userID string) {
	n.hub.RevokeChannel(workspaceID, channelID, userID)
}

// CloseWorkspaceUser はワークスペースから外されたか停止されたユーザーの接続を 4403 で切ります
func (n *Notifier) CloseWorkspaceUser(workspaceID, userID string) {
	n.hub.CloseWorkspaceUser(workspaceID, userID)
}

// CloseSession はログアウトしたセッションの接続を 4401 で切ります
func (n *Notifier) CloseSession(sessionID string) {
	n.hub.CloseSession(sessionID)
}

// CloseUser はパスワード変更やアカウント削除で全セッションを失効させたユーザーの接続を 4401 で切ります
func (n *Notifier) CloseUser(userID string) {
	n.hub.CloseUser(userID)
}

func (n *Notifier) NotifyNewMessage(workspaceID, channelID string, message messageuc.MessageOutput) {
	n.broadcastToSubscribers(workspaceID, channelID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_NewMessage{
		NewMessage: &chatv1.MessageEvent{ChannelId: channelID, Message: presenter.Message(message)},
	}})
}

func (n *Notifier) NotifyUpdatedMessage(workspaceID, channelID string, message messageuc.MessageOutput) {
	n.broadcastToSubscribers(workspaceID, channelID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_MessageUpdated{
		MessageUpdated: &chatv1.MessageEvent{ChannelId: channelID, Message: presenter.Message(message)},
	}})
}

func (n *Notifier) NotifyDeletedMessage(workspaceID, channelID string, deletion messageuc.MessageDeletion) {
	n.broadcastToSubscribers(workspaceID, channelID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_MessageDeleted{
		MessageDeleted: &chatv1.MessageDeletedEvent{
			ChannelId:         channelID,
			MessageId:         deletion.MessageID,
			DeletedMessageIds: deletion.DeletedIDs,
			DeletedAt:         timestamppb.New(deletion.DeletedAt),
		},
	}})
}

func (n *Notifier) NotifyReactionAdded(workspaceID, channelID string, reaction reactionuc.ReactionNotification) {
	n.broadcastToSubscribers(workspaceID, channelID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_ReactionAdded{
		ReactionAdded: presenter.ReactionEvent(channelID, reaction),
	}})
}

func (n *Notifier) NotifyReactionRemoved(workspaceID, channelID string, reaction reactionuc.ReactionNotification) {
	n.broadcastToSubscribers(workspaceID, channelID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_ReactionRemoved{
		ReactionRemoved: presenter.ReactionEvent(channelID, reaction),
	}})
}

func (n *Notifier) NotifyUnreadCount(workspaceID, userID, channelID string, unreadCount, mentionCount int) {
	event := &chatv1.ServerEvent{Event: &chatv1.ServerEvent_UnreadCount{
		UnreadCount: &chatv1.UnreadCountEvent{ChannelId: channelID, UnreadCount: int32(unreadCount), HasMention: mentionCount > 0, MentionCount: int32(mentionCount)},
	}}
	n.broadcastToUsers(workspaceID, []string{userID}, event)
}

// ピンの件数はチャンネルを開いていないユーザーにも表示するため、購読者ではなく参加者全員に送る

func (n *Notifier) NotifyPinCreated(workspaceID, channelID string, memberIDs []string, pin pinuc.PinNotification) {
	n.broadcastToUsers(workspaceID, memberIDs, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_PinCreated{
		PinCreated: presenter.PinEvent(channelID, pin),
	}})
}

func (n *Notifier) NotifyPinDeleted(workspaceID, channelID string, memberIDs []string, pin pinuc.PinNotification) {
	n.broadcastToUsers(workspaceID, memberIDs, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_PinDeleted{
		PinDeleted: presenter.PinEvent(channelID, pin),
	}})
}

func (n *Notifier) NotifySystemMessageCreated(workspaceID, channelID string, message *entity.SystemMessage) {
	n.broadcastToSubscribers(workspaceID, channelID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_SystemMessageCreated{
		SystemMessageCreated: &chatv1.SystemMessageEvent{
			ChannelId: channelID,
			Message:   presenter.SystemMessage(messageuc.NewSystemMessageOutput(message)),
		},
	}})
}

// 絵文字はどのチャンネルでも使うため、ワークスペースの全員に送る

func (n *Notifier) NotifyCustomEmojiCreated(workspaceID string, emoji customemojiuc.Notification) {
	n.broadcastToWorkspace(workspaceID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_CustomEmojiCreated{
		CustomEmojiCreated: &chatv1.CustomEmojiEvent{WorkspaceId: workspaceID, EmojiId: emoji.ID, Name: emoji.Name},
	}})
}

func (n *Notifier) NotifyCustomEmojiDeleted(workspaceID string, emoji customemojiuc.Notification) {
	n.broadcastToWorkspace(workspaceID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_CustomEmojiDeleted{
		CustomEmojiDeleted: &chatv1.CustomEmojiEvent{WorkspaceId: workspaceID, EmojiId: emoji.ID, Name: emoji.Name},
	}})
}

func (n *Notifier) broadcastToWorkspace(workspaceID string, event *chatv1.ServerEvent) {
	if data := encodeServerEvent(event); data != nil {
		n.hub.BroadcastToWorkspace(workspaceID, data)
	}
}
