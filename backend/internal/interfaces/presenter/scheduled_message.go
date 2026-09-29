package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
)

var scheduledMessageStatuses = map[entity.ScheduledMessageStatus]chatv1.ScheduledMessageStatus{
	entity.ScheduledMessageScheduled: chatv1.ScheduledMessageStatus_SCHEDULED_MESSAGE_STATUS_SCHEDULED,
	entity.ScheduledMessageSending:   chatv1.ScheduledMessageStatus_SCHEDULED_MESSAGE_STATUS_SENDING,
	entity.ScheduledMessageSent:      chatv1.ScheduledMessageStatus_SCHEDULED_MESSAGE_STATUS_SENT,
	entity.ScheduledMessageFailed:    chatv1.ScheduledMessageStatus_SCHEDULED_MESSAGE_STATUS_FAILED,
}

func ScheduledMessage(m *entity.ScheduledMessage) *chatv1.ScheduledMessage {
	return &chatv1.ScheduledMessage{
		Id:            m.ID,
		ChannelId:     m.ChannelID,
		ParentId:      m.ParentID,
		Body:          m.Body,
		AttachmentIds: m.AttachmentIDs,
		Location:      MessageLocation(m.Location),
		ScheduledAt:   timestamppb.New(m.ScheduledAt),
		Status:        scheduledMessageStatuses[m.Status],
		SentMessageId: m.SentMessageID,
		FailureReason: m.FailureReason,
		UpdatedAt:     timestamppb.New(m.UpdatedAt),
	}
}
