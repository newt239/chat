package repository

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/draft"
	"github.com/newt239/chat/ent/predicate"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type draftRepository struct {
	client *ent.Client
}

func NewDraftRepository(client *ent.Client) domainrepository.DraftRepository {
	return &draftRepository{client: client}
}

type draftKey struct {
	userID, channelID uuid.UUID
	parentID          *uuid.UUID
}

func parseDraftTarget(t domainrepository.DraftTarget) (draftKey, error) {
	userID, err := parseUUID(t.UserID, "user ID")
	if err != nil {
		return draftKey{}, err
	}
	channelID, err := parseUUID(t.ChannelID, "channel ID")
	if err != nil {
		return draftKey{}, err
	}
	return draftKey{userID: userID, channelID: channelID, parentID: parseUUIDPtr(t.ParentID)}, nil
}

func (k draftKey) predicate() predicate.Draft {
	parent := draft.ParentIDIsNil()
	if k.parentID != nil {
		parent = draft.ParentID(*k.parentID)
	}
	return draft.And(draft.UserID(k.userID), draft.ChannelID(k.channelID), parent)
}

func (r *draftRepository) Find(ctx context.Context, target domainrepository.DraftTarget) (*entity.Draft, error) {
	key, err := parseDraftTarget(target)
	if err != nil {
		return nil, err
	}
	d, err := orNil(transaction.ResolveClient(ctx, r.client).Draft.Query().Where(key.predicate()).Only(ctx))
	if d == nil {
		return nil, err
	}
	return draftToEntity(d), nil
}

func (r *draftRepository) FindByWorkspace(ctx context.Context, userID string, workspaceID string) ([]*entity.Draft, error) {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}
	drafts, err := transaction.ResolveClient(ctx, r.client).Draft.Query().
		Where(draft.UserID(uid), draft.HasChannelWith(channel.WorkspaceID(workspaceID))).
		Order(ent.Desc(draft.FieldUpdatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(drafts, draftToEntity), nil
}

func (r *draftRepository) Upsert(ctx context.Context, d *entity.Draft) error {
	key, err := parseDraftTarget(domainrepository.DraftTarget{UserID: d.UserID, ChannelID: d.ChannelID, ParentID: d.ParentID})
	if err != nil {
		return err
	}
	// 部分インデックスに合わせて ON CONFLICT の列と条件を選ぶ
	columns := []string{draft.FieldUserID, draft.FieldChannelID}
	where := sql.IsNull(draft.FieldParentID)
	if key.parentID != nil {
		columns = append(columns, draft.FieldParentID)
		where = sql.NotNull(draft.FieldParentID)
	}
	now := time.Now()
	id, err := transaction.ResolveClient(ctx, r.client).Draft.Create().
		SetUserID(key.userID).
		SetChannelID(key.channelID).
		SetNillableParentID(key.parentID).
		SetBody(d.Body).
		SetUpdatedAt(now).
		OnConflict(sql.ConflictColumns(columns...), sql.ConflictWhere(where)).
		UpdateBody().
		UpdateUpdatedAt().
		ID(ctx)
	if err != nil {
		return err
	}
	d.ID = id.String()
	d.UpdatedAt = now
	return nil
}

func (r *draftRepository) Delete(ctx context.Context, target domainrepository.DraftTarget) error {
	key, err := parseDraftTarget(target)
	if err != nil {
		return err
	}
	_, err = transaction.ResolveClient(ctx, r.client).Draft.Delete().Where(key.predicate()).Exec(ctx)
	return err
}

func draftToEntity(d *ent.Draft) *entity.Draft {
	return &entity.Draft{
		ID:        d.ID.String(),
		UserID:    d.UserID.String(),
		ChannelID: d.ChannelID.String(),
		ParentID:  optionalString(d.ParentID),
		Body:      d.Body,
		UpdatedAt: d.UpdatedAt,
	}
}
