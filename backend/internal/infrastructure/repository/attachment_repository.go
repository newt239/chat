package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/attachment"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/message"
	"github.com/newt239/chat/ent/user"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type attachmentRepository struct {
	client *ent.Client
}

func NewAttachmentRepository(client *ent.Client) domainrepository.AttachmentRepository {
	return &attachmentRepository{client: client}
}

func (r *attachmentRepository) FindByID(ctx context.Context, id string) (*entity.Attachment, error) {
	aid, err := utils.ParseUUID(id, "attachment ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	a, err := client.Attachment.Query().
		Where(attachment.IDEQ(aid)).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
		WithUploader().
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return utils.AttachmentToEntity(a), nil
}

func (r *attachmentRepository) Create(ctx context.Context, att *entity.Attachment) error {
	uid, err := utils.ParseUUID(att.UploaderID, "uploader ID")
	if err != nil {
		return err
	}

	cid, err := utils.ParseUUID(att.ChannelID, "channel ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	builder := client.Attachment.Create().
		SetUploaderID(uid).
		SetChannelID(cid).
		SetFileName(att.FileName).
		SetMimeType(att.MimeType).
		SetSizeBytes(att.SizeBytes).
		SetStorageKey(att.StorageKey).
		SetStatus(string(att.Status))

	if att.ID != "" {
		attachmentID, err := utils.ParseUUID(att.ID, "attachment ID")
		if err != nil {
			return err
		}
		builder = builder.SetID(attachmentID)
	}

	if att.MessageID != nil {
		mid, err := utils.ParseUUID(*att.MessageID, "message ID")
		if err != nil {
			return err
		}
		builder = builder.SetMessageID(mid)
	}

	if att.UploadedAt != nil {
		builder = builder.SetUploadedAt(*att.UploadedAt)
	}

	if att.ExpiresAt != nil {
		builder = builder.SetExpiresAt(*att.ExpiresAt)
	}

	a, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	// Load edges
	a, err = client.Attachment.Query().
		Where(attachment.IDEQ(a.ID)).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
		WithUploader().
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		Only(ctx)
	if err != nil {
		return err
	}

	*att = *utils.AttachmentToEntity(a)
	return nil
}

func (r *attachmentRepository) UpdateStatus(ctx context.Context, id string, status entity.AttachmentStatus) error {
	aid, err := utils.ParseUUID(id, "attachment ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	return client.Attachment.UpdateOneID(aid).
		SetStatus(string(status)).
		Exec(ctx)
}

func (r *attachmentRepository) CreatePending(ctx context.Context, att *entity.Attachment) error {
	aid, err := utils.ParseUUID(att.ID, "attachment ID")
	if err != nil {
		return err
	}

	uid, err := utils.ParseUUID(att.UploaderID, "uploader ID")
	if err != nil {
		return err
	}

	cid, err := utils.ParseUUID(att.ChannelID, "channel ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	a, err := client.Attachment.Create().
		SetID(aid).
		SetUploaderID(uid).
		SetChannelID(cid).
		SetFileName(att.FileName).
		SetMimeType(att.MimeType).
		SetSizeBytes(att.SizeBytes).
		SetStorageKey(att.StorageKey).
		SetStatus(string(att.Status)).
		SetUploadedAt(*att.UploadedAt).
		SetExpiresAt(*att.ExpiresAt).
		Save(ctx)
	if err != nil {
		return err
	}

	// Load edges
	a, err = client.Attachment.Query().
		Where(attachment.IDEQ(aid)).
		WithUploader().
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		Only(ctx)
	if err != nil {
		return err
	}

	*att = *utils.AttachmentToEntity(a)
	return nil
}

func (r *attachmentRepository) AttachToMessage(ctx context.Context, attachmentIDs []string, messageID string) error {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	for _, attachmentID := range attachmentIDs {
		aid, err := utils.ParseUUID(attachmentID, "attachment ID")
		if err != nil {
			return err
		}

		err = client.Attachment.UpdateOneID(aid).
			SetMessageID(mid).
			Exec(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *attachmentRepository) FindByMessageID(ctx context.Context, messageID string) ([]*entity.Attachment, error) {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	attachments, err := client.Attachment.Query().
		Where(attachment.HasMessageWith(message.ID(mid))).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
		WithUploader().
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.Attachment, 0, len(attachments))
	for _, a := range attachments {
		result = append(result, utils.AttachmentToEntity(a))
	}

	return result, nil
}

func (r *attachmentRepository) FindByMessageIDs(ctx context.Context, messageIDs []string) (map[string][]*entity.Attachment, error) {
	if len(messageIDs) == 0 {
		return make(map[string][]*entity.Attachment), nil
	}

	// Parse all message IDs
	parsedIDs := make([]uuid.UUID, 0, len(messageIDs))
	for _, id := range messageIDs {
		parsedID, err := utils.ParseUUID(id, "message ID")
		if err != nil {
			return nil, err
		}
		parsedIDs = append(parsedIDs, parsedID)
	}

	client := transaction.ResolveClient(ctx, r.client)
	attachments, err := client.Attachment.Query().
		Where(attachment.HasMessageWith(message.IDIn(parsedIDs...))).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
		WithUploader().
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]*entity.Attachment)
	for _, a := range attachments {
		messageID := a.Edges.Message.ID.String()
		if result[messageID] == nil {
			result[messageID] = make([]*entity.Attachment, 0)
		}
		result[messageID] = append(result[messageID], utils.AttachmentToEntity(a))
	}

	return result, nil
}

func (r *attachmentRepository) FindPendingByUploaderAndChannel(ctx context.Context, uploaderID, channelID string) ([]*entity.Attachment, error) {
	uid, err := utils.ParseUUID(uploaderID, "uploader ID")
	if err != nil {
		return nil, err
	}

	cid, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	attachments, err := client.Attachment.Query().
		Where(
			attachment.HasUploaderWith(user.ID(uid)),
			attachment.HasChannelWith(channel.ID(cid)),
			attachment.Status("pending"),
		).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
		WithUploader().
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.Attachment, 0, len(attachments))
	for _, a := range attachments {
		result = append(result, utils.AttachmentToEntity(a))
	}

	return result, nil
}

func (r *attachmentRepository) FindPendingByIDsForUser(ctx context.Context, userID string, attachmentIDs []string) ([]*entity.Attachment, error) {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	if len(attachmentIDs) == 0 {
		return []*entity.Attachment{}, nil
	}

	// Parse all attachment IDs
	parsedIDs := make([]uuid.UUID, 0, len(attachmentIDs))
	for _, id := range attachmentIDs {
		parsedID, err := utils.ParseUUID(id, "attachment ID")
		if err != nil {
			return nil, err
		}
		parsedIDs = append(parsedIDs, parsedID)
	}

	client := transaction.ResolveClient(ctx, r.client)
	attachments, err := client.Attachment.Query().
		Where(
			attachment.IDIn(parsedIDs...),
			attachment.HasUploaderWith(user.ID(uid)),
			attachment.Status("pending"),
		).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
		WithUploader().
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.Attachment, 0, len(attachments))
	for _, a := range attachments {
		result = append(result, utils.AttachmentToEntity(a))
	}

	return result, nil
}

func (r *attachmentRepository) Delete(ctx context.Context, id string) error {
	aid, err := utils.ParseUUID(id, "attachment ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	return client.Attachment.DeleteOneID(aid).Exec(ctx)
}
