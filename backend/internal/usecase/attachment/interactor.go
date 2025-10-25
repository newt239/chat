package attachment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/example/chat/internal/domain/entity"
	"github.com/example/chat/internal/domain/repository"
	"github.com/example/chat/internal/infrastructure/storage/wasabi"
)

type Interactor struct {
	attachmentRepo repository.AttachmentRepository
	channelRepo    repository.ChannelRepository
	messageRepo    repository.MessageRepository
	presignService *wasabi.PresignService
	config         *wasabi.Config
}

func NewInteractor(
	attachmentRepo repository.AttachmentRepository,
	channelRepo repository.ChannelRepository,
	messageRepo repository.MessageRepository,
	presignService *wasabi.PresignService,
	config *wasabi.Config,
) *Interactor {
	return &Interactor{
		attachmentRepo: attachmentRepo,
		channelRepo:    channelRepo,
		messageRepo:    messageRepo,
		presignService: presignService,
		config:         config,
	}
}

func (i *Interactor) Presign(ctx context.Context, input *PresignInput) (*PresignOutput, error) {
	if input.SizeBytes > i.config.MaxFileSize {
		return nil, fmt.Errorf("ファイルサイズが上限(1GB)を超えています")
	}

	isMember, err := i.channelRepo.IsMember(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("チャネルメンバーではありません")
	}

	attachmentID := uuid.New().String()
	storageKey := fmt.Sprintf("attachments/%s/%s", input.ChannelID, attachmentID)

	expires := time.Duration(input.ExpiresMin) * time.Minute
	if expires == 0 {
		expires = i.config.UploadExpires
	}
	expiresAt := time.Now().Add(expires)

	uploadURL, err := i.presignService.GenerateUploadURL(storageKey, input.MimeType, input.SizeBytes, expires)
	if err != nil {
		return nil, err
	}

	attachment := &entity.Attachment{
		ID:         attachmentID,
		UploaderID: input.UserID,
		ChannelID:  input.ChannelID,
		FileName:   input.FileName,
		MimeType:   input.MimeType,
		SizeBytes:  input.SizeBytes,
		StorageKey: storageKey,
		Status:     entity.AttachmentStatusPending,
		ExpiresAt:  &expiresAt,
	}

	if err := i.attachmentRepo.CreatePending(ctx, attachment); err != nil {
		return nil, err
	}

	return &PresignOutput{
		AttachmentID: attachment.ID,
		UploadURL:    uploadURL,
		StorageKey:   storageKey,
		ExpiresAt:    expiresAt,
	}, nil
}

func (i *Interactor) GetMetadata(ctx context.Context, userID, attachmentID string) (*AttachmentOutput, error) {
	attachment, err := i.attachmentRepo.FindByID(ctx, attachmentID)
	if err != nil {
		return nil, err
	}
	if attachment == nil {
		return nil, errors.New("添付ファイルが見つかりません")
	}

	if attachment.MessageID != nil {
		message, err := i.messageRepo.FindByID(ctx, *attachment.MessageID)
		if err != nil {
			return nil, err
		}
		if message == nil {
			return nil, errors.New("メッセージが見つかりません")
		}

		isMember, err := i.channelRepo.IsMember(ctx, message.ChannelID, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, errors.New("アクセス権限がありません")
		}
	} else {
		isMember, err := i.channelRepo.IsMember(ctx, attachment.ChannelID, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, errors.New("アクセス権限がありません")
		}
	}

	return &AttachmentOutput{
		ID:         attachment.ID,
		MessageID:  attachment.MessageID,
		UploaderID: attachment.UploaderID,
		ChannelID:  attachment.ChannelID,
		FileName:   attachment.FileName,
		MimeType:   attachment.MimeType,
		SizeBytes:  attachment.SizeBytes,
		Status:     string(attachment.Status),
		CreatedAt:  attachment.CreatedAt,
	}, nil
}

func (i *Interactor) GetDownloadURL(ctx context.Context, userID, attachmentID string) (*DownloadURLOutput, error) {
	attachment, err := i.attachmentRepo.FindByID(ctx, attachmentID)
	if err != nil {
		return nil, err
	}
	if attachment == nil {
		return nil, errors.New("添付ファイルが見つかりません")
	}

	if attachment.MessageID != nil {
		message, err := i.messageRepo.FindByID(ctx, *attachment.MessageID)
		if err != nil {
			return nil, err
		}
		if message == nil {
			return nil, errors.New("メッセージが見つかりません")
		}

		isMember, err := i.channelRepo.IsMember(ctx, message.ChannelID, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, errors.New("アクセス権限がありません")
		}
	} else {
		isMember, err := i.channelRepo.IsMember(ctx, attachment.ChannelID, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, errors.New("アクセス権限がありません")
		}
	}

	downloadURL, err := i.presignService.GenerateDownloadURL(attachment.StorageKey, 0)
	if err != nil {
		return nil, err
	}

	return &DownloadURLOutput{
		URL:       downloadURL,
		ExpiresIn: int(i.config.DownloadExpires.Seconds()),
	}, nil
}

func (i *Interactor) FinalizeAttachments(ctx context.Context, userID, channelID, messageID string, attachmentIDs []string) error {
	if len(attachmentIDs) == 0 {
		return nil
	}

	attachments, err := i.attachmentRepo.FindPendingByIDsForUser(ctx, userID, attachmentIDs)
	if err != nil {
		return err
	}

	if len(attachments) != len(attachmentIDs) {
		return fmt.Errorf("一部の添付ファイルが見つからないか、アクセス権限がありません")
	}

	for _, att := range attachments {
		if att.ChannelID != channelID {
			return fmt.Errorf("添付ファイルのチャネルIDが一致しません")
		}
	}

	return i.attachmentRepo.AttachToMessage(ctx, attachmentIDs, messageID)
}
