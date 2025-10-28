package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/message"
	"github.com/newt239/chat/ent/threadmetadata"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type threadRepository struct {
	client *ent.Client
}

func NewThreadRepository(client *ent.Client) domainrepository.ThreadRepository {
	return &threadRepository{client: client}
}

func (r *threadRepository) FindMetadataByMessageID(ctx context.Context, messageID string) (*entity.ThreadMetadata, error) {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	tm, err := client.ThreadMetadata.Query().
		Where(threadmetadata.HasMessageWith(message.ID(mid))).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
		WithLastReplyUser().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return utils.ThreadMetadataToEntity(tm), nil
}

func (r *threadRepository) Upsert(ctx context.Context, metadata *entity.ThreadMetadata) error {
	mid, err := utils.ParseUUID(metadata.MessageID, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	// Try to find existing
	existing, err := client.ThreadMetadata.Query().
		Where(threadmetadata.HasMessageWith(message.ID(mid))).
		Only(ctx)

	participantIDs := make([]uuid.UUID, len(metadata.ParticipantUserIDs))
	for i, id := range metadata.ParticipantUserIDs {
		participantIDs[i] = utils.ParseUUIDOrNil(id)
	}

	if err != nil {
		if ent.IsNotFound(err) {
			// Create new
			builder := client.ThreadMetadata.Create().
				SetMessageID(mid).
				SetReplyCount(metadata.ReplyCount).
				SetParticipantUserIds(participantIDs)

			if metadata.LastReplyAt != nil {
				builder = builder.SetLastReplyAt(*metadata.LastReplyAt)
			}

			if metadata.LastReplyUserID != nil {
				lruid, err := utils.ParseUUID(*metadata.LastReplyUserID, "last reply user ID")
				if err != nil {
					return err
				}
				builder = builder.SetLastReplyUserID(lruid)
			}

			_, err := builder.Save(ctx)
			if err != nil {
				return err
			}

			// Load edges
			tm, err := client.ThreadMetadata.Query().
				Where(threadmetadata.HasMessageWith(message.ID(mid))).
				WithMessage(func(q *ent.MessageQuery) {
					q.WithChannel(func(q2 *ent.ChannelQuery) {
						q2.WithWorkspace().WithCreatedBy()
					}).WithUser()
				}).
				WithLastReplyUser().
				Only(ctx)
			if err != nil {
				return err
			}

			*metadata = *utils.ThreadMetadataToEntity(tm)
			return nil
		}
		return err
	}

	// Update existing
	builder := client.ThreadMetadata.UpdateOne(existing).
		SetReplyCount(metadata.ReplyCount).
		SetParticipantUserIds(participantIDs)

	if metadata.LastReplyAt != nil {
		builder = builder.SetLastReplyAt(*metadata.LastReplyAt)
	} else {
		builder = builder.ClearLastReplyAt()
	}

	if metadata.LastReplyUserID != nil {
		lruid, err := utils.ParseUUID(*metadata.LastReplyUserID, "last reply user ID")
		if err != nil {
			return err
		}
		builder = builder.SetLastReplyUserID(lruid)
	} else {
		builder = builder.ClearLastReplyUser()
	}

	_, err = builder.Save(ctx)
	if err != nil {
		return err
	}

	// Load edges
	tm, err := client.ThreadMetadata.Query().
		Where(threadmetadata.HasMessageWith(message.ID(mid))).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
		WithLastReplyUser().
		Only(ctx)
	if err != nil {
		return err
	}

	*metadata = *utils.ThreadMetadataToEntity(tm)
	return nil
}

func (r *threadRepository) FindMetadataByMessageIDs(ctx context.Context, messageIDs []string) (map[string]*entity.ThreadMetadata, error) {
	if len(messageIDs) == 0 {
		return make(map[string]*entity.ThreadMetadata), nil
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
	metadataList, err := client.ThreadMetadata.Query().
		Where(threadmetadata.HasMessageWith(message.IDIn(parsedIDs...))).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
		WithLastReplyUser().
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*entity.ThreadMetadata)
	for _, tm := range metadataList {
		messageID := tm.Edges.Message.ID.String()
		result[messageID] = utils.ThreadMetadataToEntity(tm)
	}

	return result, nil
}

func (r *threadRepository) CreateOrUpdateMetadata(ctx context.Context, metadata *entity.ThreadMetadata) error {
	return r.Upsert(ctx, metadata)
}

func (r *threadRepository) IncrementReplyCount(ctx context.Context, messageID string, replyUserID string) error {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	// Try to find existing
	existing, err := client.ThreadMetadata.Query().
		Where(threadmetadata.HasMessageWith(message.ID(mid))).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// Create new with count 1
			_, err := client.ThreadMetadata.Create().
				SetMessageID(mid).
				SetReplyCount(1).
				SetParticipantUserIds([]uuid.UUID{}).
				Save(ctx)
			return err
		}
		return err
	}

	// Update existing
	return client.ThreadMetadata.UpdateOne(existing).
		SetReplyCount(existing.ReplyCount + 1).
		Exec(ctx)
}

func (r *threadRepository) DeleteMetadata(ctx context.Context, messageID string) error {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	_, err = client.ThreadMetadata.Delete().
		Where(threadmetadata.HasMessageWith(message.ID(mid))).
		Exec(ctx)
	return err
}
