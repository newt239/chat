package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	attachmentuc "github.com/newt239/chat/internal/usecase/attachment"
)

func Attachment(a *attachmentuc.AttachmentOutput) *chatv1.Attachment {
	return &chatv1.Attachment{
		Id:         a.ID,
		MessageId:  a.MessageID,
		UploaderId: a.UploaderID,
		ChannelId:  a.ChannelID,
		FileName:   a.FileName,
		MimeType:   a.MimeType,
		SizeBytes:  a.SizeBytes,
		Status:     a.Status,
		CreatedAt:  timestamppb.New(a.CreatedAt),
		Media:      MediaMetadata(a.Media),
	}
}
