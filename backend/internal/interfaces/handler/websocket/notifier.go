package websocket

import (
	"log"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
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
		log.Printf("[WebSocket] イベントのエンコードに失敗しました: %v", err)
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
		n.hub.BroadcastToChannelSubscribers(workspaceID, channelID, data)
	}
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
		ReactionAdded: reactionEvent(channelID, reaction),
	}})
}

func (n *Notifier) NotifyReactionRemoved(workspaceID, channelID string, reaction reactionuc.ReactionNotification) {
	n.broadcastToSubscribers(workspaceID, channelID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_ReactionRemoved{
		ReactionRemoved: reactionEvent(channelID, reaction),
	}})
}

func reactionEvent(channelID string, reaction reactionuc.ReactionNotification) *chatv1.ReactionEvent {
	event := &chatv1.ReactionEvent{ChannelId: channelID, MessageId: reaction.MessageID, UserId: reaction.UserID, Emoji: reaction.Emoji}
	if reaction.User != nil {
		event.User = &chatv1.UserSummary{Id: reaction.User.ID, DisplayName: reaction.User.DisplayName, AvatarUrl: reaction.User.AvatarURL}
		event.CreatedAt = timestamppb.New(reaction.CreatedAt)
	}
	return event
}

func (n *Notifier) NotifyUnreadCount(workspaceID, userID, channelID string, unreadCount, mentionCount int) {
	event := &chatv1.ServerEvent{Event: &chatv1.ServerEvent_UnreadCount{
		UnreadCount: &chatv1.UnreadCountEvent{ChannelId: channelID, UnreadCount: int32(unreadCount), HasMention: mentionCount > 0, MentionCount: int32(mentionCount)},
	}}
	if data := encodeServerEvent(event); data != nil {
		n.hub.BroadcastToUser(workspaceID, userID, data)
	}
}

// ピンの件数はチャンネルを開いていないユーザーにも表示するため、購読者ではなく参加者全員に送る

func (n *Notifier) NotifyPinCreated(workspaceID, channelID string, pin pinuc.PinNotification) {
	n.broadcastToChannel(workspaceID, channelID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_PinCreated{
		PinCreated: pinEvent(channelID, pin),
	}})
}

func (n *Notifier) NotifyPinDeleted(workspaceID, channelID string, pin pinuc.PinNotification) {
	n.broadcastToChannel(workspaceID, channelID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_PinDeleted{
		PinDeleted: pinEvent(channelID, pin),
	}})
}

func pinEvent(channelID string, pin pinuc.PinNotification) *chatv1.PinEvent {
	event := &chatv1.PinEvent{ChannelId: channelID, MessageId: pin.MessageID, PinnedBy: pin.PinnedBy, PinnedAt: timestamppb.New(pin.PinnedAt)}
	if pin.PinnedByUser != nil {
		event.PinnedByUser = presenter.UserSummary(*pin.PinnedByUser)
	}
	return event
}

func (n *Notifier) broadcastToChannel(workspaceID, channelID string, event *chatv1.ServerEvent) {
	if data := encodeServerEvent(event); data != nil {
		n.hub.BroadcastToChannel(workspaceID, channelID, data, "")
	}
}

func (n *Notifier) NotifySystemMessageCreated(workspaceID, channelID string, message *entity.SystemMessage) {
	n.broadcastToSubscribers(workspaceID, channelID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_SystemMessageCreated{
		SystemMessageCreated: &chatv1.SystemMessageEvent{
			ChannelId: channelID,
			Message: presenter.SystemMessage(messageuc.SystemMessageOutput{
				ID:        message.ID,
				ChannelID: message.ChannelID,
				Kind:      string(message.Kind),
				Payload:   message.Payload,
				ActorID:   message.ActorID,
				CreatedAt: message.CreatedAt,
			}),
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
