package repository

import (
	"context"

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
	workspaceID, err := utils.ParseUUID(id, "workspace ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	w, err := client.Workspace.Query().
		Where(workspace.ID(workspaceID)).
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
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
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
		SetCreatedByID(cid).
		SetName(w.Name)

	if w.ID != "" {
		wid, err := utils.ParseUUID(w.ID, "workspace ID")
		if err != nil {
			return err
		}
		builder = builder.SetID(wid)
	}

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
	wid, err := utils.ParseUUID(w.ID, "workspace ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	builder := client.Workspace.UpdateOneID(wid).
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
	wid, err := utils.ParseUUID(id, "workspace ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	return client.Workspace.DeleteOneID(wid).Exec(ctx)
}

func (r *workspaceRepository) AddMember(ctx context.Context, member *entity.WorkspaceMember) error {
	wid, err := utils.ParseUUID(member.WorkspaceID, "workspace ID")
	if err != nil {
		return err
	}

	uid, err := utils.ParseUUID(member.UserID, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	_, err = client.WorkspaceMember.Create().
		SetWorkspaceID(wid).
		SetUserID(uid).
		SetRole(string(member.Role)).
		Save(ctx)
	if err != nil {
		return err
	}

	// Load edges
	wm, err := client.WorkspaceMember.Query().
		Where(
			workspacemember.HasWorkspaceWith(workspace.ID(wid)),
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
	wid, err := utils.ParseUUID(workspaceID, "workspace ID")
	if err != nil {
		return err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	// Find the member first
	wm, err := client.WorkspaceMember.Query().
		Where(
			workspacemember.HasWorkspaceWith(workspace.ID(wid)),
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
	wid, err := utils.ParseUUID(workspaceID, "workspace ID")
	if err != nil {
		return err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	_, err = client.WorkspaceMember.Delete().
		Where(
			workspacemember.HasWorkspaceWith(workspace.ID(wid)),
			workspacemember.HasUserWith(user.ID(uid)),
		).
		Exec(ctx)

	return err
}

func (r *workspaceRepository) FindMembersByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.WorkspaceMember, error) {
	wid, err := utils.ParseUUID(workspaceID, "workspace ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	members, err := client.WorkspaceMember.Query().
		Where(workspacemember.HasWorkspaceWith(workspace.ID(wid))).
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
	wid, err := utils.ParseUUID(workspaceID, "workspace ID")
	if err != nil {
		return nil, err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	wm, err := client.WorkspaceMember.Query().
		Where(
			workspacemember.HasWorkspaceWith(workspace.ID(wid)),
			workspacemember.HasUserWith(user.ID(uid)),
		).
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
