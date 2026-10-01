package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/reminder"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
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
	creatorID, err := utils.ParseUUID(rem.CreatorID, "user ID")
	if err != nil {
		return err
	}
	created, err := transaction.ResolveClient(ctx, r.client).Reminder.Create().
		SetWorkspaceID(rem.WorkspaceID).
		SetCreatorID(creatorID).
		SetNillableTargetUserID(utils.ParseUUIDPtr(rem.TargetUserID)).
		SetNillableTargetChannelID(utils.ParseUUIDPtr(rem.TargetChannelID)).
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
	client := transaction.ResolveClient(ctx, r.client)
	rows, err := client.QueryContext(ctx, claimDueRemindersSQL, now, limit)
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
		return []*entity.Reminder{}, nil
	}
	found, err := client.Reminder.Query().Where(reminder.IDIn(ids...)).Order(ent.Asc(reminder.FieldRemindAt)).All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*entity.Reminder, 0, len(found))
	for _, rem := range found {
		result = append(result, &entity.Reminder{
			ID:              rem.ID.String(),
			WorkspaceID:     rem.WorkspaceID,
			CreatorID:       rem.CreatorID.String(),
			TargetUserID:    utils.UUIDPtrToStringPtr(rem.TargetUserID),
			TargetChannelID: utils.UUIDPtrToStringPtr(rem.TargetChannelID),
			Text:            rem.Text,
			RemindAt:        rem.RemindAt,
			CreatedAt:       rem.CreatedAt,
		})
	}
	return result, nil
}

func (r *reminderRepository) MarkSent(ctx context.Context, id string) error {
	return r.setStatus(ctx, id, reminder.StatusSent)
}

func (r *reminderRepository) MarkFailed(ctx context.Context, id string) error {
	return r.setStatus(ctx, id, reminder.StatusFailed)
}

func (r *reminderRepository) setStatus(ctx context.Context, id string, status reminder.Status) error {
	rid, err := utils.ParseUUID(id, "reminder ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).Reminder.UpdateOneID(rid).SetStatus(status).Exec(ctx)
}
