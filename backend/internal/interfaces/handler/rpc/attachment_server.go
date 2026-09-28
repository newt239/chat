package rpc

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	attachmentuc "github.com/newt239/chat/internal/usecase/attachment"
)

type AttachmentServer struct {
	UC *attachmentuc.Interactor
}

// 添付ファイルのユースケースは専用のエラー値を持たないため、失敗の種類をここで決める

func (s *AttachmentServer) PresignUpload(ctx context.Context, req *chatv1.PresignUploadRequest) (*chatv1.PresignUploadResponse, error) {
	out, err := s.UC.Presign(ctx, &attachmentuc.PresignInput{
		UserID:    userIDFrom(ctx),
		ChannelID: req.ChannelId,
		FileName:  req.FileName,
		MimeType:  req.ContentType,
		SizeBytes: req.SizeBytes,
		Media:     entity.MediaMetadata{Width: req.Width, Height: req.Height, DurationSeconds: req.DurationSeconds},
		Thumbnail: thumbnailInput(req.Thumbnail),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return &chatv1.PresignUploadResponse{
		AttachmentId:       out.AttachmentID,
		UploadUrl:          out.UploadURL,
		ThumbnailUploadUrl: out.ThumbnailUploadURL,
		ExpiresAt:          timestamppb.New(out.ExpiresAt),
	}, nil
}

func thumbnailInput(t *chatv1.ThumbnailUpload) *attachmentuc.ThumbnailInput {
	if t == nil {
		return nil
	}
	return &attachmentuc.ThumbnailInput{MimeType: t.ContentType, SizeBytes: t.SizeBytes, Width: t.Width, Height: t.Height}
}

func (s *AttachmentServer) GetAttachment(ctx context.Context, req *chatv1.GetAttachmentRequest) (*chatv1.GetAttachmentResponse, error) {
	out, err := s.UC.GetMetadata(ctx, userIDFrom(ctx), req.AttachmentId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return &chatv1.GetAttachmentResponse{Attachment: &chatv1.Attachment{
		Id:         out.ID,
		MessageId:  out.MessageID,
		UploaderId: out.UploaderID,
		ChannelId:  out.ChannelID,
		FileName:   out.FileName,
		MimeType:   out.MimeType,
		SizeBytes:  out.SizeBytes,
		Status:     out.Status,
		CreatedAt:  timestamppb.New(out.CreatedAt),
		Media:      presenter.MediaMetadata(out.Media),
	}}, nil
}

func (s *AttachmentServer) GetDownloadUrl(ctx context.Context, req *chatv1.GetDownloadUrlRequest) (*chatv1.GetDownloadUrlResponse, error) {
	out, err := s.UC.GetDownloadURL(ctx, userIDFrom(ctx), req.AttachmentId, req.Thumbnail)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return &chatv1.GetDownloadUrlResponse{Url: out.URL, ExpiresIn: int32(out.ExpiresIn)}, nil
}

func (s *AttachmentServer) DeleteAttachment(ctx context.Context, req *chatv1.DeleteAttachmentRequest) (*chatv1.DeleteAttachmentResponse, error) {
	if err := s.UC.Delete(ctx, userIDFrom(ctx), req.AttachmentId); err != nil {
		return nil, err
	}
	return &chatv1.DeleteAttachmentResponse{}, nil
}
