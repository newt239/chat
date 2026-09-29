package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/scheduledmessage"
	"github.com/newt239/chat/ent/workspace"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

// 期限の来た予約を他のワーカーがロック中のものを飛ばして取り出し、同じ文で送信中にする
const claimDueSQL = `
	UPDATE scheduled_message SET status = 'sending', updated_at = $1
	WHERE id IN (
		SELECT id FROM scheduled_message
		WHERE status = 'scheduled' AND scheduled_at <= $1
		ORDER BY scheduled_at
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	)
	RETURNING id`

type scheduledMessageRepository struct {
	client *ent.Client
}

func NewScheduledMessageRepository(client *ent.Client) domainrepository.ScheduledMessageRepository {
	return &scheduledMessageRepository{client: client}
}

func (r *scheduledMessageRepository) Create(ctx context.Context, m *entity.ScheduledMessage) error {
	userID, err := utils.ParseUUID(m.UserID, "user ID")
	if err != nil {
		return err
	}
	channelID, err := utils.ParseUUID(m.ChannelID, "channel ID")
	if err != nil {
		return err
	}
	builder := transaction.ResolveClient(ctx, r.client).ScheduledMessage.Create().
		SetUserID(userID).
		SetChannelID(channelID).
		SetBody(m.Body).
		SetAttachmentIds(m.AttachmentIDs).
		SetScheduledAt(m.ScheduledAt)
	if m.ParentID != nil {
		parentID, err := utils.ParseUUID(*m.ParentID, "parent ID")
		if err != nil {
			return err
		}
		builder.SetParentID(parentID)
	}
	if loc := m.Location; loc != nil {
		builder.
			SetLocationLatitude(loc.Latitude).
			SetLocationLongitude(loc.Longitude).
			SetNillableLocationAccuracy(loc.AccuracyMeters).
			SetNillableLocationLabel(loc.Label)
	}
	saved, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	*m = *scheduledMessageToEntity(saved)
	return nil
}

func (r *scheduledMessageRepository) FindByID(ctx context.Context, id string) (*entity.ScheduledMessage, error) {
	sid, err := utils.ParseUUID(id, "scheduled message ID")
	if err != nil {
		return nil, err
	}
	m, err := transaction.ResolveClient(ctx, r.client).ScheduledMessage.Get(ctx, sid)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return scheduledMessageToEntity(m), nil
}

func (r *scheduledMessageRepository) FindByWorkspace(ctx context.Context, userID string, workspaceID string) ([]*entity.ScheduledMessage, error) {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}
	messages, err := transaction.ResolveClient(ctx, r.client).ScheduledMessage.Query().
		Where(
			scheduledmessage.UserID(uid),
			scheduledmessage.HasChannelWith(channel.HasWorkspaceWith(workspace.ID(workspaceID))),
		).
		Order(ent.Asc(scheduledmessage.FieldScheduledAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return scheduledMessagesToEntities(messages), nil
}

func (r *scheduledMessageRepository) Reschedule(ctx context.Context, id string, body string, scheduledAt time.Time) error {
	sid, err := utils.ParseUUID(id, "scheduled message ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).ScheduledMessage.UpdateOneID(sid).
		SetBody(body).
		SetScheduledAt(scheduledAt).
		SetStatus(scheduledmessage.StatusScheduled).
		ClearFailureReason().
		Exec(ctx)
}

func (r *scheduledMessageRepository) Delete(ctx context.Context, id string) error {
	sid, err := utils.ParseUUID(id, "scheduled message ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).ScheduledMessage.DeleteOneID(sid).Exec(ctx)
}

func (r *scheduledMessageRepository) ClaimDue(ctx context.Context, now time.Time, limit int) ([]*entity.ScheduledMessage, error) {
	client := transaction.ResolveClient(ctx, r.client)
	rows, err := client.QueryContext(ctx, claimDueSQL, now, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	ids := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []*entity.ScheduledMessage{}, nil
	}
	messages, err := client.ScheduledMessage.Query().
		Where(scheduledmessage.IDIn(ids...)).
		Order(ent.Asc(scheduledmessage.FieldScheduledAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return scheduledMessagesToEntities(messages), nil
}

func (r *scheduledMessageRepository) Claim(ctx context.Context, id string) (*entity.ScheduledMessage, error) {
	sid, err := utils.ParseUUID(id, "scheduled message ID")
	if err != nil {
		return nil, err
	}
	claimed, err := transaction.ResolveClient(ctx, r.client).ScheduledMessage.Update().
		Where(
			scheduledmessage.ID(sid),
			scheduledmessage.StatusIn(scheduledmessage.StatusScheduled, scheduledmessage.StatusFailed),
		).
		SetStatus(scheduledmessage.StatusSending).
		ClearFailureReason().
		Save(ctx)
	if err != nil || claimed == 0 {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *scheduledMessageRepository) MarkSent(ctx context.Context, id string, messageID string) error {
	sid, err := utils.ParseUUID(id, "scheduled message ID")
	if err != nil {
		return err
	}
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).ScheduledMessage.UpdateOneID(sid).
		SetStatus(scheduledmessage.StatusSent).
		SetSentMessageID(mid).
		Exec(ctx)
}

func (r *scheduledMessageRepository) MarkFailed(ctx context.Context, id string, reason string) error {
	sid, err := utils.ParseUUID(id, "scheduled message ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).ScheduledMessage.UpdateOneID(sid).
		SetStatus(scheduledmessage.StatusFailed).
		SetFailureReason(reason).
		Exec(ctx)
}

func scheduledMessagesToEntities(messages []*ent.ScheduledMessage) []*entity.ScheduledMessage {
	result := make([]*entity.ScheduledMessage, 0, len(messages))
	for _, m := range messages {
		result = append(result, scheduledMessageToEntity(m))
	}
	return result
}

func scheduledMessageToEntity(m *ent.ScheduledMessage) *entity.ScheduledMessage {
	result := &entity.ScheduledMessage{
		ID:            m.ID.String(),
		UserID:        m.UserID.String(),
		ChannelID:     m.ChannelID.String(),
		Body:          m.Body,
		AttachmentIDs: m.AttachmentIds,
		Location:      utils.LocationToEntity(m.LocationLatitude, m.LocationLongitude, m.LocationAccuracy, m.LocationLabel),
		ScheduledAt:   m.ScheduledAt,
		Status:        entity.ScheduledMessageStatus(m.Status),
		FailureReason: m.FailureReason,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
	if result.AttachmentIDs == nil {
		result.AttachmentIDs = []string{}
	}
	if m.ParentID != nil {
		pid := m.ParentID.String()
		result.ParentID = &pid
	}
	if m.SentMessageID != nil {
		sid := m.SentMessageID.String()
		result.SentMessageID = &sid
	}
	return result
}
