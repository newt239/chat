package websocket

import (
	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	pinuc "github.com/newt239/chat/internal/usecase/pin"
	reactionuc "github.com/newt239/chat/internal/usecase/reaction"
)

// Notifier は各ユースケースの変更通知を WebSocket のイベントとして配信します。接続を切る操作は Hub のものをそのまま使う
type Notifier struct {
	*Hub
}

func (n Notifier) NotifyNewMessage(workspaceID, channelID string, message messageuc.MessageOutput) {
	n.BroadcastToChannel(workspaceID, channelID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_NewMessage{
		NewMessage: &chatv1.MessageEvent{ChannelId: channelID, Message: presenter.Message(message)},
	}})
}

func (n Notifier) NotifyUpdatedMessage(workspaceID, channelID string, message messageuc.MessageOutput) {
	n.BroadcastToChannel(workspaceID, channelID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_MessageUpdated{
		MessageUpdated: &chatv1.MessageEvent{ChannelId: channelID, Message: presenter.Message(message)},
	}})
}

func (n Notifier) NotifyDeletedMessage(workspaceID, channelID string, deletion messageuc.MessageDeletion) {
	n.BroadcastToChannel(workspaceID, channelID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_MessageDeleted{
		MessageDeleted: &chatv1.MessageDeletedEvent{
			ChannelId:         channelID,
			MessageId:         deletion.MessageID,
			DeletedMessageIds: deletion.DeletedIDs,
		},
	}})
}

func (n Notifier) NotifySystemMessageCreated(workspaceID, channelID string, message *entity.SystemMessage) {
	n.BroadcastToChannel(workspaceID, channelID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_SystemMessageCreated{
		SystemMessageCreated: &chatv1.SystemMessageEvent{ChannelId: channelID, Message: presenter.SystemMessage(message)},
	}})
}

func (n Notifier) NotifyReactionAdded(workspaceID, channelID string, reaction reactionuc.ReactionNotification) {
	n.BroadcastToChannel(workspaceID, channelID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_ReactionAdded{
		ReactionAdded: presenter.ReactionEvent(channelID, reaction),
	}})
}

func (n Notifier) NotifyReactionRemoved(workspaceID, channelID string, reaction reactionuc.ReactionNotification) {
	n.BroadcastToChannel(workspaceID, channelID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_ReactionRemoved{
		ReactionRemoved: presenter.ReactionEvent(channelID, reaction),
	}})
}

func (n Notifier) NotifyUnreadCount(workspaceID, userID, channelID string, unreadCount, mentionCount int) {
	n.BroadcastToUsers(workspaceID, []string{userID}, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_UnreadCount{
		UnreadCount: &chatv1.UnreadCountEvent{ChannelId: channelID, UnreadCount: int32(unreadCount), MentionCount: int32(mentionCount)},
	}})
}

// ピンの件数はチャンネルを開いていないユーザーにも表示するため、購読者ではなく参加者全員に送る

func (n Notifier) NotifyPinCreated(workspaceID, channelID string, memberIDs []string, pin pinuc.PinNotification) {
	n.BroadcastToUsers(workspaceID, memberIDs, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_PinCreated{
		PinCreated: presenter.PinEvent(channelID, pin),
	}})
}

func (n Notifier) NotifyPinDeleted(workspaceID, channelID string, memberIDs []string, pin pinuc.PinNotification) {
	n.BroadcastToUsers(workspaceID, memberIDs, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_PinDeleted{
		PinDeleted: presenter.PinEvent(channelID, pin),
	}})
}

// 絵文字はどのチャンネルでも使うため、ワークスペースの全員に送る
func (n Notifier) NotifyCustomEmojisChanged(workspaceID string) {
	n.BroadcastToWorkspace(workspaceID, &chatv1.ServerEvent{Event: &chatv1.ServerEvent_CustomEmojisChanged{
		CustomEmojisChanged: &chatv1.CustomEmojisChangedEvent{},
	}})
}
