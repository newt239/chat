package attachment

import (
	"context"
	"fmt"
	"mime"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	"github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

var (
	ErrThumbnailNotAllowed = domerr.New(domerr.ErrValidation, "サムネイルは動画にだけ付けられます")
	ErrThumbnailNotFound   = domerr.New(domerr.ErrNotFound, "サムネイルがありません")
	ErrFileTooLarge        = domerr.New(domerr.ErrValidation, "ファイルサイズが上限を超えています")
)

type PresignInput struct {
	UserID    string
	ChannelID string
	FileName  string
	MimeType  string
	SizeBytes int64
	Media     entity.MediaMetadata
	Thumbnail *ThumbnailInput
}

// ThumbnailInput は動画と一緒にアップロードするサムネイル画像です
type ThumbnailInput struct {
	MimeType  string
	SizeBytes int64
	Width     int32
	Height    int32
}

type PresignOutput struct {
	AttachmentID string
	UploadURL    string
	// サムネイルを指定したときだけ返します
	ThumbnailUploadURL *string
}

type Interactor struct {
	attachmentRepo   repository.AttachmentRepository
	messageRepo      repository.MessageRepository
	channelAccessSvc service.ChannelAccessService
	storageService   service.StorageService
}

func New(
	attachmentRepo repository.AttachmentRepository,
	messageRepo repository.MessageRepository,
	channelAccessSvc service.ChannelAccessService,
	storageService service.StorageService,
) *Interactor {
	return &Interactor{
		attachmentRepo:   attachmentRepo,
		messageRepo:      messageRepo,
		channelAccessSvc: channelAccessSvc,
		storageService:   storageService,
	}
}

func (i *Interactor) Presign(ctx context.Context, input PresignInput) (*PresignOutput, error) {
	if input.SizeBytes > service.MaxUploadSize {
		return nil, ErrFileTooLarge
	}
	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID); err != nil {
		return nil, err
	}

	attachmentID := uuid.NewString()
	storageKey := fmt.Sprintf("attachments/%s/%s", input.ChannelID, attachmentID)
	mimeType := normalizeMimeType(input.MimeType, input.FileName)
	uploadURL, err := i.storageService.GenerateUploadURL(ctx, storageKey, mimeType, input.SizeBytes, service.UploadURLExpires)
	if err != nil {
		return nil, err
	}

	media := input.Media
	var thumbnailUploadURL *string
	if t := input.Thumbnail; t != nil {
		if !strings.HasPrefix(mimeType, "video/") {
			return nil, ErrThumbnailNotAllowed
		}
		thumbnailKey := storageKey + "-thumbnail"
		url, err := i.storageService.GenerateUploadURL(ctx, thumbnailKey, t.MimeType, t.SizeBytes, service.UploadURLExpires)
		if err != nil {
			return nil, err
		}
		thumbnailUploadURL = &url
		media.Thumbnail = &entity.Thumbnail{StorageKey: thumbnailKey, Width: t.Width, Height: t.Height}
	}

	if err := i.attachmentRepo.CreatePending(ctx, &entity.Attachment{
		ID:         attachmentID,
		UploaderID: input.UserID,
		ChannelID:  input.ChannelID,
		FileName:   input.FileName,
		MimeType:   mimeType,
		SizeBytes:  input.SizeBytes,
		Media:      media,
		StorageKey: storageKey,
	}); err != nil {
		return nil, err
	}
	return &PresignOutput{
		AttachmentID:       attachmentID,
		UploadURL:          uploadURL,
		ThumbnailUploadURL: thumbnailUploadURL,
	}, nil
}

// GetDownloadURL は閲覧できるチャンネルの添付の、thumbnail が true のときはサムネイル画像の署名付き URL を返します
func (i *Interactor) GetDownloadURL(ctx context.Context, userID, attachmentID string, thumbnail bool) (string, error) {
	attachment, err := i.attachmentRepo.FindByID(ctx, attachmentID)
	if err != nil {
		return "", err
	}
	if attachment == nil {
		return "", domerr.ErrAttachmentNotFound
	}
	channelID := attachment.ChannelID
	if attachment.MessageID != nil {
		message, err := i.messageRepo.FindByID(ctx, *attachment.MessageID)
		if err != nil {
			return "", err
		}
		if message == nil {
			return "", domerr.ErrMessageNotFound
		}
		channelID = message.ChannelID
	}
	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, channelID, userID); err != nil {
		return "", err
	}

	storageKey := attachment.StorageKey
	if thumbnail {
		if attachment.Media.Thumbnail == nil {
			return "", ErrThumbnailNotFound
		}
		storageKey = attachment.Media.Thumbnail.StorageKey
	}
	return i.storageService.GenerateDownloadURL(ctx, storageKey, service.DownloadURLExpires)
}

// normalizeMimeType はパラメータを除いて小文字にし、判別できない種別はファイル名の拡張子から推定します
func normalizeMimeType(mimeType, fileName string) string {
	mediaType, _, err := mime.ParseMediaType(mimeType)
	if err == nil && mediaType != "application/octet-stream" {
		return mediaType
	}
	if guessed, _, err := mime.ParseMediaType(mime.TypeByExtension(filepath.Ext(fileName))); err == nil {
		return guessed
	}
	return "application/octet-stream"
}
