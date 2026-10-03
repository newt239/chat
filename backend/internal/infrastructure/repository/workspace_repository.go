package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/predicate"
	"github.com/newt239/chat/ent/user"
	"github.com/newt239/chat/ent/workspace"
	"github.com/newt239/chat/ent/workspacemember"
	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type workspaceRepository struct {
	client *ent.Client
}

func NewWorkspaceRepository(client *ent.Client) domainrepository.WorkspaceRepository {
	return &workspaceRepository{client: client}
}

func (r *workspaceRepository) query(ctx context.Context) *ent.WorkspaceQuery {
	return transaction.ResolveClient(ctx, r.client).Workspace.Query()
}

func (r *workspaceRepository) members(ctx context.Context) *ent.WorkspaceMemberQuery {
	return transaction.ResolveClient(ctx, r.client).WorkspaceMember.Query()
}

func (r *workspaceRepository) FindByID(ctx context.Context, id string) (*entity.Workspace, error) {
	w, err := orNil(r.query(ctx).Where(workspace.ID(id)).Only(ctx))
	if w == nil {
		return nil, err
	}
	return workspaceToEntity(w), nil
}

func (r *workspaceRepository) Create(ctx context.Context, w *entity.Workspace) error {
	createdBy, err := parseUUID(w.CreatedBy, "created by user ID")
	if err != nil {
		return err
	}
	ws, err := transaction.ResolveClient(ctx, r.client).Workspace.Create().
		SetID(w.ID).
		SetCreatedByID(createdBy).
		SetName(w.Name).
		SetNillableDescription(w.Description).
		SetNillableIconURL(w.IconURL).
		SetIsPublic(w.IsPublic).
		SetSignupEnabled(w.SignupEnabled).
		SetEmailSignupEnabled(w.EmailSignupEnabled).
		Save(ctx)
	if ent.IsConstraintError(err) {
		return domerr.ErrWorkspaceIDExists
	}
	if err != nil {
		return err
	}
	*w = *workspaceToEntity(ws)
	return nil
}

func (r *workspaceRepository) Update(ctx context.Context, w *entity.Workspace) error {
	builder := transaction.ResolveClient(ctx, r.client).Workspace.UpdateOneID(w.ID).
		SetName(w.Name).
		SetIsPublic(w.IsPublic).
		SetSignupEnabled(w.SignupEnabled).
		SetEmailSignupEnabled(w.EmailSignupEnabled)
	if w.Description != nil {
		builder.SetDescription(*w.Description)
	} else {
		builder.ClearDescription()
	}
	if w.IconURL != nil {
		builder.SetIconURL(*w.IconURL)
	} else {
		builder.ClearIconURL()
	}
	ws, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	*w = *workspaceToEntity(ws)
	return nil
}

func (r *workspaceRepository) Delete(ctx context.Context, id string) error {
	return transaction.ResolveClient(ctx, r.client).Workspace.DeleteOneID(id).Exec(ctx)
}

func (r *workspaceRepository) AddMember(ctx context.Context, member *entity.WorkspaceMember) error {
	uid, err := parseUUID(member.UserID, "user ID")
	if err != nil {
		return err
	}
	wm, err := transaction.ResolveClient(ctx, r.client).WorkspaceMember.Create().
		SetWorkspaceID(member.WorkspaceID).
		SetUserID(uid).
		SetRole(string(member.Role)).
		Save(ctx)
	if err != nil {
		return err
	}
	*member = *workspaceMemberToEntity(wm)
	return nil
}

func memberWhere(workspaceID, userID string) (predicate.WorkspaceMember, error) {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}
	return workspacemember.And(workspacemember.WorkspaceID(workspaceID), workspacemember.UserID(uid)), nil
}

func (r *workspaceRepository) UpdateMemberRole(ctx context.Context, workspaceID string, userID string, role entity.WorkspaceRole) error {
	where, err := memberWhere(workspaceID, userID)
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).WorkspaceMember.Update().Where(where).SetRole(string(role)).Exec(ctx)
}

func (r *workspaceRepository) SetMemberSuspended(ctx context.Context, workspaceID string, userID string, suspendedAt *time.Time) error {
	where, err := memberWhere(workspaceID, userID)
	if err != nil {
		return err
	}
	update := transaction.ResolveClient(ctx, r.client).WorkspaceMember.Update().Where(where)
	if suspendedAt == nil {
		update.ClearSuspendedAt()
	} else {
		update.SetSuspendedAt(*suspendedAt)
	}
	return update.Exec(ctx)
}

func (r *workspaceRepository) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	where, err := memberWhere(workspaceID, userID)
	if err != nil {
		return err
	}
	_, err = transaction.ResolveClient(ctx, r.client).WorkspaceMember.Delete().Where(where).Exec(ctx)
	return err
}

func (r *workspaceRepository) FindMembersByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.WorkspaceMember, error) {
	members, err := r.members(ctx).Where(workspacemember.WorkspaceID(workspaceID)).All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(members, workspaceMemberToEntity), nil
}

func (r *workspaceRepository) FindMember(ctx context.Context, workspaceID string, userID string) (*entity.WorkspaceMember, error) {
	return r.findMember(ctx, workspaceID, userID, workspacemember.SuspendedAtIsNil())
}

func (r *workspaceRepository) FindMemberIncludingSuspended(ctx context.Context, workspaceID string, userID string) (*entity.WorkspaceMember, error) {
	return r.findMember(ctx, workspaceID, userID)
}

func (r *workspaceRepository) findMember(ctx context.Context, workspaceID string, userID string, predicates ...predicate.WorkspaceMember) (*entity.WorkspaceMember, error) {
	where, err := memberWhere(workspaceID, userID)
	if err != nil {
		return nil, err
	}
	wm, err := orNil(r.members(ctx).Where(where).Where(predicates...).Only(ctx))
	if wm == nil {
		return nil, err
	}
	return workspaceMemberToEntity(wm), nil
}

func (r *workspaceRepository) SearchMembers(ctx context.Context, workspaceID string, query string, limit int, offset int) ([]*entity.WorkspaceMember, int, error) {
	memberQuery := r.members(ctx).Where(workspacemember.WorkspaceID(workspaceID))
	if keyword := strings.TrimSpace(query); keyword != "" {
		memberQuery.Where(workspacemember.HasUserWith(user.Or(user.DisplayNameContainsFold(keyword), user.EmailContainsFold(keyword))))
	}
	total, err := memberQuery.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	members, err := memberQuery.Offset(offset).Limit(limit).Order(ent.Asc(workspacemember.FieldJoinedAt)).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return convertAll(members, workspaceMemberToEntity), total, nil
}

func (r *workspaceRepository) FindAllPublic(ctx context.Context) ([]*entity.Workspace, error) {
	list, err := r.query(ctx).Where(workspace.IsPublic(true)).Order(ent.Desc(workspace.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(list, workspaceToEntity), nil
}

// CountMembersBatch は停止中も含めたメンバー数をワークスペースごとに返します
func (r *workspaceRepository) CountMembersBatch(ctx context.Context, workspaceIDs []string) (map[string]int, error) {
	var rows []struct {
		WorkspaceID string `json:"workspace_id"`
		Count       int    `json:"count"`
	}
	err := r.members(ctx).
		Where(workspacemember.WorkspaceIDIn(workspaceIDs...)).
		GroupBy(workspacemember.FieldWorkspaceID).
		Aggregate(ent.Count()).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int, len(rows))
	for _, row := range rows {
		counts[row.WorkspaceID] = row.Count
	}
	return counts, nil
}

func (r *workspaceRepository) FindMembershipsByUserID(ctx context.Context, userID string) ([]*entity.WorkspaceMember, error) {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}
	members, err := r.members(ctx).Where(workspacemember.UserID(uid), workspacemember.SuspendedAtIsNil()).WithWorkspace().All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(members, func(wm *ent.WorkspaceMember) *entity.WorkspaceMember {
		m := workspaceMemberToEntity(wm)
		m.Workspace = workspaceToEntity(wm.Edges.Workspace)
		return m
	}), nil
}

func (r *workspaceRepository) FindActiveMemberIDs(ctx context.Context, workspaceID string, userIDs []string) (map[string]bool, error) {
	uids, err := parseUUIDs(userIDs, "user ID")
	if err != nil {
		return nil, err
	}
	members, err := r.members(ctx).
		Where(workspacemember.WorkspaceID(workspaceID), workspacemember.UserIDIn(uids...), workspacemember.SuspendedAtIsNil()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return idSet(members, func(m *ent.WorkspaceMember) uuid.UUID { return m.UserID }), nil
}

// $1: ワークスペース ID, $2: 投稿数を数え始める日時
const memberActivitiesSQL = `
	SELECT wm.user_id,
		(SELECT COUNT(*) FROM message m JOIN channel c ON c.id = m.channel_id
			WHERE c.workspace_id = $1 AND m.user_id = wm.user_id AND m.deleted_at IS NULL AND m.created_at >= $2),
		(SELECT COALESCE(SUM(a.size_bytes), 0) FROM attachment a JOIN channel c ON c.id = a.channel_id
			WHERE c.workspace_id = $1 AND a.uploader_id = wm.user_id AND a.status = 'attached'),
		(SELECT MAX(m.created_at) FROM message m JOIN channel c ON c.id = m.channel_id
			WHERE c.workspace_id = $1 AND m.user_id = wm.user_id AND m.deleted_at IS NULL)
	FROM workspace_member wm WHERE wm.workspace_id = $1`

func (r *workspaceRepository) FindMemberActivities(ctx context.Context, workspaceID string, since time.Time) (map[string]entity.MemberActivity, error) {
	rows, err := transaction.ResolveClient(ctx, r.client).QueryContext(ctx, memberActivitiesSQL, workspaceID, since)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := map[string]entity.MemberActivity{}
	for rows.Next() {
		var userID uuid.UUID
		var a entity.MemberActivity
		var last sql.NullTime
		if err := rows.Scan(&userID, &a.MessageCount, &a.StorageBytes, &last); err != nil {
			return nil, err
		}
		if last.Valid {
			a.LastMessageAt = &last.Time
		}
		result[userID.String()] = a
	}
	return result, rows.Err()
}
