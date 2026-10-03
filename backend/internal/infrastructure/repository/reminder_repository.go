package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/reminder"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

// 期限の来たリマインダーを他のワーカーがロック中のものを飛ばして取り出し、同じ文で送信中にする
const claimDueRemindersSQL = `
	UPDATE reminder SET status = 'sending', updated_at = $1
	WHERE id IN (
		SELECT id FROM reminder
		WHERE status = 'scheduled' AND remind_at <= $1
		ORDER BY remind_at
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	)
	RETURNING id`

type reminderRepository struct {
	client *ent.Client
}

func NewReminderRepository(client *ent.Client) domainrepository.ReminderRepository {
	return &reminderRepository{client: client}
}

func (r *reminderRepository) Create(ctx context.Context, rem *entity.Reminder) error {
	creatorID, err := parseUUID(rem.CreatorID, "user ID")
	if err != nil {
		return err
	}
	created, err := transaction.ResolveClient(ctx, r.client).Reminder.Create().
		SetWorkspaceID(rem.WorkspaceID).
		SetCreatorID(creatorID).
		SetNillableTargetUserID(parseUUIDPtr(rem.TargetUserID)).
		SetNillableTargetChannelID(parseUUIDPtr(rem.TargetChannelID)).
		SetText(rem.Text).
		SetRemindAt(rem.RemindAt).
		Save(ctx)
	if err != nil {
		return err
	}
	rem.ID = created.ID.String()
	rem.CreatedAt = created.CreatedAt
	return nil
}

func (r *reminderRepository) ClaimDue(ctx context.Context, now time.Time, limit int) ([]*entity.Reminder, error) {
	ids, err := queryUUIDs(ctx, r.client, claimDueRemindersSQL, now, limit)
	if err != nil {
		return nil, err
	}
	found, err := transaction.ResolveClient(ctx, r.client).Reminder.Query().
		Where(reminder.IDIn(ids...)).
		Order(ent.Asc(reminder.FieldRemindAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(found, func(rem *ent.Reminder) *entity.Reminder {
		return &entity.Reminder{
			ID:              rem.ID.String(),
			WorkspaceID:     rem.WorkspaceID,
			CreatorID:       rem.CreatorID.String(),
			TargetUserID:    optionalString(rem.TargetUserID),
			TargetChannelID: optionalString(rem.TargetChannelID),
			Text:            rem.Text,
			RemindAt:        rem.RemindAt,
			CreatedAt:       rem.CreatedAt,
		}
	}), nil
}

func (r *reminderRepository) MarkSent(ctx context.Context, id string) error {
	return r.setStatus(ctx, id, reminder.StatusSent)
}

func (r *reminderRepository) MarkFailed(ctx context.Context, id string) error {
	return r.setStatus(ctx, id, reminder.StatusFailed)
}

func (r *reminderRepository) setStatus(ctx context.Context, id string, status reminder.Status) error {
	rid, err := parseUUID(id, "reminder ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).Reminder.UpdateOneID(rid).SetStatus(status).Exec(ctx)
}
