package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/attachment"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type attachmentRepository struct {
	client *ent.Client
}

func NewAttachmentRepository(client *ent.Client) domainrepository.AttachmentRepository {
	return &attachmentRepository{client: client}
}

func (r *attachmentRepository) FindByID(ctx context.Context, id string) (*entity.Attachment, error) {
	aid, err := parseUUID(id, "attachment ID")
	if err != nil {
		return nil, err
	}
	a, err := orNil(transaction.ResolveClient(ctx, r.client).Attachment.Get(ctx, aid))
	if a == nil {
		return nil, err
	}
	return attachmentToEntity(a), nil
}

func (r *attachmentRepository) CreatePending(ctx context.Context, att *entity.Attachment) error {
	aid, err := parseUUID(att.ID, "attachment ID")
	if err != nil {
		return err
	}
	uid, err := parseUUID(att.UploaderID, "uploader ID")
	if err != nil {
		return err
	}
	cid, err := parseUUID(att.ChannelID, "channel ID")
	if err != nil {
		return err
	}
	create := transaction.ResolveClient(ctx, r.client).Attachment.Create().
		SetID(aid).
		SetUploaderID(uid).
		SetChannelID(cid).
		SetFileName(att.FileName).
		SetMimeType(att.MimeType).
		SetSizeBytes(att.SizeBytes).
		SetStorageKey(att.StorageKey).
		SetStatus(string(entity.AttachmentStatusPending)).
		SetNillableWidth(att.Media.Width).
		SetNillableHeight(att.Media.Height).
		SetNillableDurationSeconds(att.Media.DurationSeconds)
	if t := att.Media.Thumbnail; t != nil {
		create.SetThumbnailStorageKey(t.StorageKey).SetThumbnailWidth(t.Width).SetThumbnailHeight(t.Height)
	}
	a, err := create.Save(ctx)
	if err != nil {
		return err
	}
	*att = *attachmentToEntity(a)
	return nil
}

func (r *attachmentRepository) AttachToMessage(ctx context.Context, attachmentIDs []string, messageID string) error {
	mid, err := parseUUID(messageID, "message ID")
	if err != nil {
		return err
	}
	aids, err := parseUUIDs(attachmentIDs, "attachment ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).Attachment.Update().
		Where(attachment.IDIn(aids...)).
		SetMessageID(mid).
		SetStatus(string(entity.AttachmentStatusAttached)).
		Exec(ctx)
}

func (r *attachmentRepository) FindByMessageIDs(ctx context.Context, messageIDs []string) (map[string][]*entity.Attachment, error) {
	ids, err := parseUUIDs(messageIDs, "message ID")
	if err != nil {
		return nil, err
	}
	attachments, err := transaction.ResolveClient(ctx, r.client).Attachment.Query().
		Where(attachment.MessageIDIn(ids...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string][]*entity.Attachment)
	for _, a := range attachments {
		messageID := a.MessageID.String()
		result[messageID] = append(result[messageID], attachmentToEntity(a))
	}
	return result, nil
}

func (r *attachmentRepository) FindPendingByIDsForUser(ctx context.Context, userID string, attachmentIDs []string) ([]*entity.Attachment, error) {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}
	ids, err := parseUUIDs(attachmentIDs, "attachment ID")
	if err != nil {
		return nil, err
	}
	attachments, err := transaction.ResolveClient(ctx, r.client).Attachment.Query().
		Where(
			attachment.IDIn(ids...),
			attachment.UploaderID(uid),
			attachment.Status(string(entity.AttachmentStatusPending)),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(attachments, attachmentToEntity), nil
}
