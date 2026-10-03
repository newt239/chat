package rpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/domain/service"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	attachmentuc "github.com/newt239/chat/internal/usecase/attachment"
)

type AttachmentServer struct {
	UC *attachmentuc.Interactor
}

func (s *AttachmentServer) PresignUpload(ctx context.Context, req *chatv1.PresignUploadRequest) (*chatv1.PresignUploadResponse, error) {
	input := attachmentuc.PresignInput{
		UserID:    userIDFrom(ctx),
		ChannelID: req.ChannelId,
		FileName:  req.FileName,
		MimeType:  req.ContentType,
		SizeBytes: req.SizeBytes,
		Media:     entity.MediaMetadata{Width: req.Width, Height: req.Height, DurationSeconds: req.DurationSeconds},
	}
	if t := req.Thumbnail; t != nil {
		input.Thumbnail = &attachmentuc.ThumbnailInput{MimeType: t.ContentType, Width: t.Width, Height: t.Height}
	}
	out, err := s.UC.Presign(ctx, input)
	if err != nil {
		return nil, err
	}
	return &chatv1.PresignUploadResponse{
		AttachmentId:       out.AttachmentID,
		UploadUrl:          out.UploadURL,
		ThumbnailUploadUrl: out.ThumbnailUploadURL,
		ExpiresAt:          timestamppb.New(out.ExpiresAt),
	}, nil
}

func (s *AttachmentServer) GetDownloadUrl(ctx context.Context, req *chatv1.GetDownloadUrlRequest) (*chatv1.GetDownloadUrlResponse, error) {
	url, err := s.UC.GetDownloadURL(ctx, userIDFrom(ctx), req.AttachmentId, req.Thumbnail)
	if err != nil {
		return nil, err
	}
	return &chatv1.GetDownloadUrlResponse{Url: url, ExpiresIn: int32(service.DownloadURLExpires.Seconds())}, nil
}
