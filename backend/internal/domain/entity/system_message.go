package entity

import "time"

type SystemMessageKind string

const (
    SystemMessageKindMemberJoined            SystemMessageKind = "member_joined"
    SystemMessageKindMemberAdded             SystemMessageKind = "member_added"
    SystemMessageKindChannelPrivacyChanged   SystemMessageKind = "channel_privacy_changed"
    SystemMessageKindChannelNameChanged      SystemMessageKind = "channel_name_changed"
    SystemMessageKindChannelDescriptionChanged SystemMessageKind = "channel_description_changed"
    SystemMessageKindMessagePinned           SystemMessageKind = "message_pinned"
)

type SystemMessage struct {
    ID        string
    ChannelID string
    Kind      SystemMessageKind
    Payload   map[string]any
    ActorID   *string
    CreatedAt time.Time
}


