package attachment

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	"github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

var (
	ErrThumbnailNotAllowed = errors.New("サムネイルは動画にだけ付けられます")
	ErrThumbnailNotFound   = errors.New("サムネイルがありません")
	ErrFileTooLarge        = fmt.Errorf("%w: ファイルサイズが上限を超えています", domerr.ErrValidation)
)

type Interactor struct {
	attachmentRepo   repository.AttachmentRepository
	messageRepo      repository.MessageRepository
	channelAccessSvc service.ChannelAccessService
	storageService   service.StorageService
	config           service.StorageConfig
}

func NewInteractor(
	attachmentRepo repository.AttachmentRepository,
	messageRepo repository.MessageRepository,
	channelAccessSvc service.ChannelAccessService,
	storageService service.StorageService,
	config service.StorageConfig,
) *Interactor {
	return &Interactor{
		attachmentRepo:   attachmentRepo,
		messageRepo:      messageRepo,
		channelAccessSvc: channelAccessSvc,
		storageService:   storageService,
		config:           config,
	}
}

func (i *Interactor) Presign(ctx context.Context, input *PresignInput) (*PresignOutput, error) {
	if input.SizeBytes > i.config.GetMaxFileSize() {
		return nil, ErrFileTooLarge
	}

	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID); err != nil {
		return nil, err
	}

	attachmentID := uuid.New().String()
	storageKey := fmt.Sprintf("attachments/%s/%s", input.ChannelID, attachmentID)

	expires := time.Duration(input.ExpiresMin) * time.Minute
	if expires == 0 {
		expires = i.config.GetUploadExpires().(time.Duration)
	}
	expiresAt := time.Now().Add(expires)

	mimeType := normalizeMimeType(input.MimeType, input.FileName)
	uploadURL, err := i.storageService.GenerateUploadURL(storageKey, mimeType, input.SizeBytes, expires)
	if err != nil {
		return nil, err
	}

	media := input.Media
	var thumbnailUploadURL *string
	if input.Thumbnail != nil {
		if !strings.HasPrefix(mimeType, "video/") {
			return nil, ErrThumbnailNotAllowed
		}
		thumbnailKey := storageKey + "-thumbnail"
		url, err := i.storageService.GenerateUploadURL(thumbnailKey, input.Thumbnail.MimeType, input.Thumbnail.SizeBytes, expires)
		if err != nil {
			return nil, err
		}
		thumbnailUploadURL = &url
		media.Thumbnail = &entity.Thumbnail{StorageKey: thumbnailKey, Width: input.Thumbnail.Width, Height: input.Thumbnail.Height}
	}

	attachment := &entity.Attachment{
		ID:         attachmentID,
		UploaderID: input.UserID,
		ChannelID:  input.ChannelID,
		FileName:   input.FileName,
		MimeType:   mimeType,
		SizeBytes:  input.SizeBytes,
		Media:      media,
		StorageKey: storageKey,
		Status:     entity.AttachmentStatusPending,
		ExpiresAt:  &expiresAt,
	}

	if err := i.attachmentRepo.CreatePending(ctx, attachment); err != nil {
		return nil, err
	}

	return &PresignOutput{
		AttachmentID:       attachment.ID,
		UploadURL:          uploadURL,
		ThumbnailUploadURL: thumbnailUploadURL,
		StorageKey:         storageKey,
		ExpiresAt:          expiresAt,
	}, nil
}

// findAccessible は閲覧者が参照できるチャンネルの添付だけを返します
func (i *Interactor) findAccessible(ctx context.Context, userID, attachmentID string) (*entity.Attachment, error) {
	attachment, err := i.attachmentRepo.FindByID(ctx, attachmentID)
	if err != nil {
		return nil, err
	}
	if attachment == nil {
		return nil, domerr.ErrAttachmentNotFound
	}

	channelID := attachment.ChannelID
	if attachment.MessageID != nil {
		message, err := i.messageRepo.FindByID(ctx, *attachment.MessageID)
		if err != nil {
			return nil, err
		}
		if message == nil {
			return nil, domerr.ErrMessageNotFound
		}
		channelID = message.ChannelID
	}
	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, channelID, userID); err != nil {
		return nil, err
	}
	return attachment, nil
}

func (i *Interactor) GetMetadata(ctx context.Context, userID, attachmentID string) (*AttachmentOutput, error) {
	attachment, err := i.findAccessible(ctx, userID, attachmentID)
	if err != nil {
		return nil, err
	}

	return &AttachmentOutput{
		ID:         attachment.ID,
		MessageID:  attachment.MessageID,
		UploaderID: attachment.UploaderID,
		ChannelID:  attachment.ChannelID,
		FileName:   attachment.FileName,
		MimeType:   attachment.MimeType,
		SizeBytes:  attachment.SizeBytes,
		Media:      attachment.Media,
		Status:     string(attachment.Status),
		CreatedAt:  attachment.CreatedAt,
	}, nil
}

// GetDownloadURL は本体、thumbnail が true のときはサムネイル画像の署名付き URL を返します
func (i *Interactor) GetDownloadURL(ctx context.Context, userID, attachmentID string, thumbnail bool) (*DownloadURLOutput, error) {
	attachment, err := i.findAccessible(ctx, userID, attachmentID)
	if err != nil {
		return nil, err
	}

	storageKey := attachment.StorageKey
	if thumbnail {
		if attachment.Media.Thumbnail == nil {
			return nil, ErrThumbnailNotFound
		}
		storageKey = attachment.Media.Thumbnail.StorageKey
	}
	downloadURL, err := i.storageService.GenerateDownloadURL(storageKey, 0)
	if err != nil {
		return nil, err
	}

	return &DownloadURLOutput{
		URL:       downloadURL,
		ExpiresIn: int(i.config.GetDownloadExpires().(time.Duration).Seconds()),
	}, nil
}

// Delete は添付ファイルを削除します。削除できるのはアップロードした本人のみです
func (i *Interactor) Delete(ctx context.Context, userID, attachmentID string) error {
	attachment, err := i.attachmentRepo.FindByID(ctx, attachmentID)
	if err != nil {
		return err
	}
	if attachment == nil {
		return domerr.ErrAttachmentNotFound
	}
	if attachment.UploaderID != userID {
		return domerr.ErrUnauthorized
	}

	if err := i.attachmentRepo.Delete(ctx, attachmentID); err != nil {
		return err
	}

	if t := attachment.Media.Thumbnail; t != nil {
		if err := i.storageService.DeleteObject(t.StorageKey); err != nil {
			return err
		}
	}
	return i.storageService.DeleteObject(attachment.StorageKey)
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
