package presenter

import (
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

func UserSummary(u messageuc.UserInfo) *chatv1.UserSummary {
	return &chatv1.UserSummary{Id: u.ID, DisplayName: u.DisplayName, AvatarUrl: u.AvatarURL, IsBot: u.IsBot}
}

func Message(m messageuc.MessageOutput) *chatv1.Message {
	msg := &chatv1.Message{
		Id:        m.ID,
		ChannelId: m.ChannelID,
		UserId:    m.UserID,
		User:      UserSummary(m.User),
		ParentId:  m.ParentID,
		Body:      m.Body,
		Mentions: ConvertAll(m.Mentions, func(u messageuc.UserMention) *chatv1.UserMention {
			return &chatv1.UserMention{UserId: u.UserID, DisplayName: u.DisplayName}
		}),
		Groups: ConvertAll(m.Groups, func(g messageuc.GroupMention) *chatv1.GroupMention {
			return &chatv1.GroupMention{GroupId: g.GroupID, Name: g.Name}
		}),
		Links: ConvertAll(m.Links, func(l messageuc.LinkInfo) *chatv1.MessageLink {
			return &chatv1.MessageLink{
				Id:              l.ID,
				Url:             l.URL,
				Ogp:             OGPData(l.OGP),
				LinkedMessageId: l.LinkedMessageID,
				MessagePreview:  messagePreview(l.MessagePreview),
			}
		}),
		Reactions: ConvertAll(m.Reactions, func(r messageuc.ReactionInfo) *chatv1.Reaction {
			return &chatv1.Reaction{MessageId: m.ID, User: UserSummary(r.User), Emoji: r.Emoji, CreatedAt: timestamppb.New(r.CreatedAt)}
		}),
		Attachments: ConvertAll(m.Attachments, func(a messageuc.AttachmentInfo) *chatv1.MessageAttachment {
			return &chatv1.MessageAttachment{Id: a.ID, FileName: a.FileName, MimeType: a.MimeType, SizeBytes: a.SizeBytes, Media: MediaMetadata(a.Media)}
		}),
		CreatedAt: timestamppb.New(m.CreatedAt),
		EditedAt:  optionalTimestamp(m.EditedAt),
		DeletedAt: optionalTimestamp(m.DeletedAt),
		IsDeleted: m.IsDeleted,
	}
	if m.DeletedBy != nil {
		msg.DeletedBy = UserSummary(*m.DeletedBy)
	}
	if m.Pin != nil {
		msg.Pin = &chatv1.MessagePin{PinnedBy: UserSummary(m.Pin.PinnedBy), PinnedAt: timestamppb.New(m.Pin.PinnedAt)}
	}
	msg.Location = MessageLocation(m.Location)
	return msg
}

func MessageLocation(l *entity.MessageLocation) *chatv1.MessageLocation {
	if l == nil {
		return nil
	}
	return &chatv1.MessageLocation{Latitude: l.Latitude, Longitude: l.Longitude, AccuracyMeters: l.AccuracyMeters, Label: l.Label}
}

func OGPData(o entity.OGPData) *chatv1.OgpData {
	data := &chatv1.OgpData{
		Title:       o.Title,
		Description: o.Description,
		ImageUrl:    o.ImageURL,
		SiteName:    o.SiteName,
		CardType:    o.CardType,
		ImageWidth:  o.ImageWidth,
		ImageHeight: o.ImageHeight,
	}
	if o.YouTube != nil {
		data.Youtube = &chatv1.YouTubeVideo{VideoId: o.YouTube.VideoID, ChannelName: o.YouTube.ChannelName, DurationSeconds: o.YouTube.DurationSeconds}
	}
	return data
}

func MediaMetadata(m entity.MediaMetadata) *chatv1.MediaMetadata {
	out := &chatv1.MediaMetadata{Width: m.Width, Height: m.Height, DurationSeconds: m.DurationSeconds}
	if m.Thumbnail != nil {
		out.Thumbnail = &chatv1.MediaThumbnail{Width: m.Thumbnail.Width, Height: m.Thumbnail.Height}
	}
	return out
}

func MessagePreview(p messageuc.MessagePreviewOutput) *chatv1.MessagePreview {
	return &chatv1.MessagePreview{
		MessageId:   p.MessageID,
		ChannelId:   p.ChannelID,
		ChannelName: p.ChannelName,
		ParentId:    p.ParentID,
		User:        UserSummary(p.User),
		BodyExcerpt: p.BodyExcerpt,
		CreatedAt:   timestamppb.New(p.CreatedAt),
	}
}

func messagePreview(p *messageuc.MessagePreviewOutput) *chatv1.MessagePreview {
	if p == nil {
		return nil
	}
	return MessagePreview(*p)
}

func MessageWithThread(m messageuc.MessageWithThreadOutput) *chatv1.Message {
	msg := Message(m.MessageOutput)
	if m.ThreadMetadata != nil {
		msg.ThreadMetadata = ThreadMetadata(*m.ThreadMetadata)
	}
	return msg
}

func ThreadMetadata(t messageuc.ThreadMetadataOutput) *chatv1.ThreadMetadata {
	metadata := &chatv1.ThreadMetadata{
		MessageId:          t.MessageID,
		ReplyCount:         int32(t.ReplyCount),
		LastReplyAt:        optionalTimestamp(t.LastReplyAt),
		ParticipantUserIds: t.ParticipantUserIDs,
	}
	if t.LastReplyUser != nil {
		metadata.LastReplyUser = UserSummary(*t.LastReplyUser)
	}
	return metadata
}

func SystemMessage(s messageuc.SystemMessageOutput) *chatv1.SystemMessage {
	// payload は文字列など JSON で表せる値だけを持つため変換に失敗しない
	payload, _ := structpb.NewStruct(s.Payload)
	return &chatv1.SystemMessage{
		Id:        s.ID,
		ChannelId: s.ChannelID,
		Kind:      chatv1.SystemMessageKind(chatv1.SystemMessageKind_value["SYSTEM_MESSAGE_KIND_"+strings.ToUpper(s.Kind)]),
		Payload:   payload,
		ActorId:   s.ActorID,
		CreatedAt: timestamppb.New(s.CreatedAt),
	}
}

func TimelineItem(i messageuc.TimelineItem) *chatv1.TimelineItem {
	item := &chatv1.TimelineItem{CreatedAt: timestamppb.New(i.CreatedAt)}
	switch {
	case i.UserMessage != nil:
		item.Content = &chatv1.TimelineItem_UserMessage{UserMessage: Message(*i.UserMessage)}
	case i.SystemMessage != nil:
		item.Content = &chatv1.TimelineItem_SystemMessage{SystemMessage: SystemMessage(*i.SystemMessage)}
	}
	return item
}
