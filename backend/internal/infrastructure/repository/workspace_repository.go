package repository

import (
    "context"
    "strings"

    "entgo.io/ent/dialect/sql"
    "github.com/newt239/chat/ent"
    "github.com/newt239/chat/ent/user"
    "github.com/newt239/chat/ent/workspace"
    "github.com/newt239/chat/ent/workspacemember"
    "github.com/newt239/chat/internal/domain/entity"
    domainrepository "github.com/newt239/chat/internal/domain/repository"
    "github.com/newt239/chat/internal/infrastructure/transaction"
    "github.com/newt239/chat/internal/infrastructure/utils"
)

type workspaceRepository struct {
    client *ent.Client
}

func NewWorkspaceRepository(client *ent.Client) domainrepository.WorkspaceRepository {
	return &workspaceRepository{client: client}
}

func (r *workspaceRepository) FindByID(ctx context.Context, id string) (*entity.Workspace, error) {
    client := transaction.ResolveClient(ctx, r.client)
    w, err := client.Workspace.Query().
        Where(workspace.ID(id)).
        WithCreatedBy().
        Only(ctx)
    if err != nil {
        if ent.IsNotFound(err) {
            return nil, nil
        }
        return nil, err
    }

    return utils.WorkspaceToEntity(w), nil
}

func (r *workspaceRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.Workspace, error) {
	client := transaction.ResolveClient(ctx, r.client)
    uid, err := utils.ParseUUID(userID, "user ID")
    if err != nil {
        return nil, err
    }
	workspaces, err := client.Workspace.Query().
        Where(workspace.HasMembersWith(workspacemember.HasUserWith(user.ID(uid)))).
		WithCreatedBy().
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.Workspace, 0, len(workspaces))
	for _, w := range workspaces {
		result = append(result, utils.WorkspaceToEntity(w))
	}

	return result, nil
}

func (r *workspaceRepository) Create(ctx context.Context, w *entity.Workspace) error {
    cid, err := utils.ParseUUID(w.CreatedBy, "created by user ID")
    if err != nil {
        return err
    }

    client := transaction.ResolveClient(ctx, r.client)

    builder := client.Workspace.Create().
        SetID(w.ID).
        SetCreatedByID(cid).
        SetName(w.Name).
        SetIsPublic(w.IsPublic)

    if w.Description != nil {
        builder = builder.SetDescription(*w.Description)
    }

    if w.IconURL != nil {
        builder = builder.SetIconURL(*w.IconURL)
    }

    ws, err := builder.Save(ctx)
    if err != nil {
        return err
    }

    // Load edges
    ws, err = client.Workspace.Query().
        Where(workspace.ID(ws.ID)).
        WithCreatedBy().
        Only(ctx)
    if err != nil {
        return err
    }

    *w = *utils.WorkspaceToEntity(ws)
    return nil
}

func (r *workspaceRepository) Update(ctx context.Context, w *entity.Workspace) error {
    client := transaction.ResolveClient(ctx, r.client)

    builder := client.Workspace.UpdateOneID(w.ID).
        SetName(w.Name)

    if w.Description != nil {
        builder = builder.SetDescription(*w.Description)
    } else {
        builder = builder.ClearDescription()
    }

    if w.IconURL != nil {
        builder = builder.SetIconURL(*w.IconURL)
    } else {
        builder = builder.ClearIconURL()
    }

    builder = builder.SetIsPublic(w.IsPublic)

    ws, err := builder.Save(ctx)
    if err != nil {
        return err
    }

    // Load edges
    ws, err = client.Workspace.Query().
        Where(workspace.ID(ws.ID)).
        WithCreatedBy().
        Only(ctx)
    if err != nil {
        return err
    }

    *w = *utils.WorkspaceToEntity(ws)
    return nil
}

func (r *workspaceRepository) Delete(ctx context.Context, id string) error {
    client := transaction.ResolveClient(ctx, r.client)
    return client.Workspace.DeleteOneID(id).Exec(ctx)
}

func (r *workspaceRepository) AddMember(ctx context.Context, member *entity.WorkspaceMember) error {
    uid, err := utils.ParseUUID(member.UserID, "user ID")
    if err != nil {
        return err
    }

    client := transaction.ResolveClient(ctx, r.client)

    _, err = client.WorkspaceMember.Create().
        SetWorkspaceID(member.WorkspaceID).
        SetUserID(uid).
        SetRole(string(member.Role)).
        Save(ctx)
    if err != nil {
        return err
    }

    // Load edges
    wm, err := client.WorkspaceMember.Query().
        Where(
            workspacemember.HasWorkspaceWith(workspace.ID(member.WorkspaceID)),
            workspacemember.HasUserWith(user.ID(uid)),
        ).
        WithWorkspace().
        WithUser().
        Only(ctx)
    if err != nil {
        return err
    }

    *member = *utils.WorkspaceMemberToEntity(wm)
    return nil
}

func (r *workspaceRepository) UpdateMemberRole(ctx context.Context, workspaceID string, userID string, role entity.WorkspaceRole) error {
    uid, err := utils.ParseUUID(userID, "user ID")
    if err != nil {
        return err
    }

    client := transaction.ResolveClient(ctx, r.client)

    // Find the member first
    wm, err := client.WorkspaceMember.Query().
        Where(
            workspacemember.HasWorkspaceWith(workspace.ID(workspaceID)),
            workspacemember.HasUserWith(user.ID(uid)),
        ).
        Only(ctx)
    if err != nil {
        return err
    }

    // Update the role
    _, err = client.WorkspaceMember.UpdateOne(wm).
        SetRole(string(role)).
        Save(ctx)

    return err
}

func (r *workspaceRepository) RemoveMember(ctx context.Context, workspaceID, userID string) error {
    uid, err := utils.ParseUUID(userID, "user ID")
    if err != nil {
        return err
    }

    client := transaction.ResolveClient(ctx, r.client)
    _, err = client.WorkspaceMember.Delete().
        Where(
            workspacemember.HasWorkspaceWith(workspace.ID(workspaceID)),
            workspacemember.HasUserWith(user.ID(uid)),
        ).
        Exec(ctx)

    return err
}

func (r *workspaceRepository) FindMembersByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.WorkspaceMember, error) {
    client := transaction.ResolveClient(ctx, r.client)
    members, err := client.WorkspaceMember.Query().
        Where(workspacemember.HasWorkspaceWith(workspace.ID(workspaceID))).
        WithWorkspace().
        WithUser().
        All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.WorkspaceMember, 0, len(members))
	for _, wm := range members {
		result = append(result, utils.WorkspaceMemberToEntity(wm))
	}

	return result, nil
}

func (r *workspaceRepository) FindMember(ctx context.Context, workspaceID string, userID string) (*entity.WorkspaceMember, error) {
    uid, err := utils.ParseUUID(userID, "user ID")
    if err != nil {
        return nil, err
    }

    client := transaction.ResolveClient(ctx, r.client)
    wm, err := client.WorkspaceMember.Query().
        Where(func(s *sql.Selector) {
            s.Where(sql.EQ(workspacemember.WorkspaceColumn, workspaceID))
            s.Where(sql.EQ(workspacemember.UserColumn, uid))
        }).
        WithWorkspace().
        WithUser().
        Only(ctx)
	if err != nil {
        if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return utils.WorkspaceMemberToEntity(wm), nil
}

func (r *workspaceRepository) SearchMembers(ctx context.Context, workspaceID string, query string, limit int, offset int) ([]*entity.WorkspaceMember, int, error) {
    client := transaction.ResolveClient(ctx, r.client)
	trimmedQuery := strings.TrimSpace(query)

    memberQuery := client.WorkspaceMember.Query().
        Where(workspacemember.HasWorkspaceWith(workspace.ID(workspaceID))).
		WithWorkspace().
		WithUser()

	if trimmedQuery != "" {
		memberQuery = memberQuery.Where(
			workspacemember.HasUserWith(
				user.Or(
					user.DisplayNameContainsFold(trimmedQuery),
					user.EmailContainsFold(trimmedQuery),
				),
			),
		)
	}

    total, err := memberQuery.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	if offset > 0 {
		memberQuery = memberQuery.Offset(offset)
	}

	if limit > 0 {
		memberQuery = memberQuery.Limit(limit)
	}

    members, err := memberQuery.
        Order(ent.Asc(workspacemember.FieldJoinedAt)).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*entity.WorkspaceMember, 0, len(members))
	for _, wm := range members {
		result = append(result, utils.WorkspaceMemberToEntity(wm))
	}

	return result, total, nil
}

// New methods for slug-based workspace
func (r *workspaceRepository) FindAllPublic(ctx context.Context) ([]*entity.Workspace, error) {
    client := transaction.ResolveClient(ctx, r.client)
    list, err := client.Workspace.Query().
        Where(workspace.IsPublic(true)).
        WithCreatedBy().
        Order(ent.Desc(workspace.FieldCreatedAt)).
        All(ctx)
    if err != nil {
        return nil, err
    }

    result := make([]*entity.Workspace, 0, len(list))
    for _, w := range list {
        result = append(result, utils.WorkspaceToEntity(w))
    }
    return result, nil
}

func (r *workspaceRepository) CountMembers(ctx context.Context, workspaceID string) (int, error) {
    client := transaction.ResolveClient(ctx, r.client)
    count, err := client.WorkspaceMember.Query().
        Where(workspacemember.HasWorkspaceWith(workspace.ID(workspaceID))).
        Count(ctx)
    if err != nil {
        return 0, err
    }
    return count, nil
}

func (r *workspaceRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
    client := transaction.ResolveClient(ctx, r.client)
    count, err := client.Workspace.Query().
        Where(workspace.ID(id)).
        Count(ctx)
    if err != nil {
        return false, err
    }
    return count > 0, nil
}
