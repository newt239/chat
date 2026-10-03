package workspace

import (
	"context"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	domaintransaction "github.com/newt239/chat/internal/domain/transaction"
)

var (
	ErrWorkspaceNotPublic = domerr.New(domerr.ErrUnauthorized, "このワークスペースは公開されていません")
)

type Interactor struct {
	workspaceRepo domainrepository.WorkspaceRepository
	userRepo      domainrepository.UserRepository
	userNoteRepo  domainrepository.UserNoteRepository
	txManager     domaintransaction.Manager
}

func New(
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
	userNoteRepo domainrepository.UserNoteRepository,
	txManager domaintransaction.Manager,
) *Interactor {
	return &Interactor{
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
		userNoteRepo:  userNoteRepo,
		txManager:     txManager,
	}
}

func (i *Interactor) ListWorkspaces(ctx context.Context, userID string) ([]WorkspaceOutput, error) {
	workspaces, err := i.workspaceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspaces: %w", err)
	}
	memberships, err := i.workspaceRepo.FindMembershipsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get memberships: %w", err)
	}
	roles := make(map[string]entity.WorkspaceRole, len(memberships))
	for _, m := range memberships {
		roles[m.WorkspaceID] = m.Role
	}
	outputs := make([]WorkspaceOutput, 0, len(workspaces))
	for _, ws := range workspaces {
		if role, ok := roles[ws.ID]; ok {
			outputs = append(outputs, WorkspaceOutput{Workspace: ws, Role: role})
		}
	}
	return outputs, nil
}

func (i *Interactor) GetWorkspace(ctx context.Context, workspaceID, userID string) (*WorkspaceOutput, error) {
	member, err := domainservice.EnsureMember(ctx, i.workspaceRepo, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	ws, err := i.findWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	return &WorkspaceOutput{Workspace: ws, Role: member.Role}, nil
}

func (i *Interactor) findWorkspace(ctx context.Context, workspaceID string) (*entity.Workspace, error) {
	ws, err := i.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	if ws == nil {
		return nil, domerr.ErrWorkspaceNotFound
	}
	return ws, nil
}

// CreateWorkspace は ID が使われていれば ErrWorkspaceIDExists を返します
func (i *Interactor) CreateWorkspace(ctx context.Context, input CreateWorkspaceInput) (*WorkspaceOutput, error) {
	if err := entity.ValidateWorkspaceSlug(input.ID); err != nil {
		return nil, err
	}
	workspace := &entity.Workspace{
		ID:          input.ID,
		Name:        input.Name,
		Description: input.Description,
		IconURL:     input.IconURL,
		IsPublic:    input.IsPublic,
		CreatedBy:   input.CreatedBy,
	}
	err := i.txManager.Do(ctx, func(txCtx context.Context) error {
		if err := i.workspaceRepo.Create(txCtx, workspace); err != nil {
			return err
		}
		return i.workspaceRepo.AddMember(txCtx, &entity.WorkspaceMember{WorkspaceID: workspace.ID, UserID: input.CreatedBy, Role: entity.WorkspaceRoleOwner})
	})
	if err != nil {
		return nil, err
	}
	return &WorkspaceOutput{Workspace: workspace, Role: entity.WorkspaceRoleOwner}, nil
}

func (i *Interactor) UpdateWorkspace(ctx context.Context, input UpdateWorkspaceInput) (*WorkspaceOutput, error) {
	member, err := domainservice.EnsureAdmin(ctx, i.workspaceRepo, input.ID, input.UserID)
	if err != nil {
		return nil, err
	}
	ws, err := i.findWorkspace(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		ws.Name = *input.Name
	}
	if input.Description != nil {
		ws.Description = input.Description
	}
	// 空文字はアイコンを外す
	if input.IconURL != nil {
		ws.IconURL = input.IconURL
		if *input.IconURL == "" {
			ws.IconURL = nil
		}
	}
	if input.IsPublic != nil {
		ws.IsPublic = *input.IsPublic
	}
	if input.SignupEnabled != nil {
		ws.SignupEnabled = *input.SignupEnabled
	}
	if input.EmailSignupEnabled != nil {
		ws.EmailSignupEnabled = *input.EmailSignupEnabled
	}
	if err := i.workspaceRepo.Update(ctx, ws); err != nil {
		return nil, fmt.Errorf("failed to update workspace: %w", err)
	}
	return &WorkspaceOutput{Workspace: ws, Role: member.Role}, nil
}

// DeleteWorkspace はオーナーだけが実行できます
func (i *Interactor) DeleteWorkspace(ctx context.Context, workspaceID, userID string) error {
	member, err := domainservice.EnsureMember(ctx, i.workspaceRepo, workspaceID, userID)
	if err != nil {
		return err
	}
	if member.Role != entity.WorkspaceRoleOwner {
		return domerr.ErrUnauthorized
	}
	if err := i.workspaceRepo.Delete(ctx, workspaceID); err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}
	return nil
}

func (i *Interactor) ListMembers(ctx context.Context, workspaceID, requesterID string) ([]MemberInfo, error) {
	if _, err := domainservice.EnsureMember(ctx, i.workspaceRepo, workspaceID, requesterID); err != nil {
		return nil, err
	}
	members, err := i.workspaceRepo.FindMembersByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}
	userIDs := make([]string, len(members))
	for idx, m := range members {
		userIDs[idx] = m.UserID
	}
	users, err := i.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	nicknames, err := i.userNoteRepo.FindNicknames(ctx, requesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get nicknames: %w", err)
	}
	infos := NewMemberInfos(members, users)
	for idx := range infos {
		if nickname, ok := nicknames[infos[idx].UserID]; ok {
			infos[idx].Nickname = &nickname
		}
	}
	return infos, nil
}

// ListPublicWorkspaces は公開ワークスペースをメンバー数と参加済みかを付けて返します
func (i *Interactor) ListPublicWorkspaces(ctx context.Context, userID string) ([]PublicWorkspaceItem, error) {
	workspaces, err := i.workspaceRepo.FindAllPublic(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list public workspaces: %w", err)
	}
	joined, err := i.workspaceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user workspaces: %w", err)
	}
	joinedMap := make(map[string]bool, len(joined))
	for _, w := range joined {
		joinedMap[w.ID] = true
	}
	ids := make([]string, len(workspaces))
	for idx, w := range workspaces {
		ids[idx] = w.ID
	}
	counts, err := i.workspaceRepo.CountMembersBatch(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to count members: %w", err)
	}
	items := make([]PublicWorkspaceItem, 0, len(workspaces))
	for _, w := range workspaces {
		items = append(items, PublicWorkspaceItem{Workspace: w, MemberCount: counts[w.ID], IsJoined: joinedMap[w.ID]})
	}
	return items, nil
}

func (i *Interactor) JoinPublicWorkspace(ctx context.Context, workspaceID, userID string) error {
	ws, err := i.findWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}
	if !ws.IsPublic && !ws.SignupEnabled {
		return ErrWorkspaceNotPublic
	}
	// 停止中のメンバーが参加し直して停止を解かないよう、停止中も参加済みとして扱う
	existing, err := i.workspaceRepo.FindMemberIncludingSuspended(ctx, workspaceID, userID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if existing != nil {
		return domerr.ErrAlreadyMember
	}
	if err := i.workspaceRepo.AddMember(ctx, &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: userID, Role: entity.WorkspaceRoleMember}); err != nil {
		return fmt.Errorf("failed to join workspace: %w", err)
	}
	return nil
}

// GetSignupInfo は参加リンクの画面に出す情報を返します。登録を許可していなければ存在しないものとして扱います
func (i *Interactor) GetSignupInfo(ctx context.Context, workspaceID string) (*entity.Workspace, error) {
	ws, err := i.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	if ws == nil || !ws.SignupEnabled {
		return nil, domerr.ErrWorkspaceNotFound
	}
	return ws, nil
}
