package workspace

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/audit"
)

var (
	ErrWorkspaceNotFound     = errors.New("ワークスペースが見つかりません")
	ErrUnauthorized          = errors.New("この操作を行う権限がありません")
	ErrInvalidRole           = errors.New("無効なワークスペースロールです")
	ErrCannotRemoveOwner     = errors.New("ワークスペースのオーナーは削除できません")
	ErrCannotChangeOwnerRole = errors.New("オーナーのロールは変更できません")
)

type WorkspaceUseCase interface {
	GetWorkspacesByUserID(ctx context.Context, userID string) (*GetWorkspacesOutput, error)
	GetWorkspace(ctx context.Context, input GetWorkspaceInput) (*GetWorkspaceOutput, error)
	CreateWorkspace(ctx context.Context, input CreateWorkspaceInput) (*CreateWorkspaceOutput, error)
	UpdateWorkspace(ctx context.Context, input UpdateWorkspaceInput) (*UpdateWorkspaceOutput, error)
	DeleteWorkspace(ctx context.Context, input DeleteWorkspaceInput) (*DeleteWorkspaceOutput, error)
	ListMembers(ctx context.Context, input ListMembersInput) (*ListMembersOutput, error)
	UpdateMemberRole(ctx context.Context, input UpdateMemberRoleInput) (*MemberActionOutput, error)
	RemoveMember(ctx context.Context, input RemoveMemberInput) (*MemberActionOutput, error)

	// 新規
	ListPublicWorkspaces(ctx context.Context, userID string) (*ListPublicWorkspacesOutput, error)
	JoinPublicWorkspace(ctx context.Context, input JoinPublicWorkspaceInput) (*MemberActionOutput, error)
	GetSignupInfo(ctx context.Context, workspaceID string) (*SignupInfoOutput, error)
}

// MemberCloser はワークスペースから外したメンバーのリアルタイム接続を切ります
type MemberCloser interface {
	CloseWorkspaceUser(workspaceID, userID string)
}

type workspaceInteractor struct {
	workspaceRepo domainrepository.WorkspaceRepository
	userRepo      domainrepository.UserRepository
	userNoteRepo  domainrepository.UserNoteRepository
	permissionSvc domainservice.PermissionService
	recorder      audit.Recorder
	memberCloser  MemberCloser
}

func NewWorkspaceInteractor(
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
	userNoteRepo domainrepository.UserNoteRepository,
	permissionSvc domainservice.PermissionService,
	recorder audit.Recorder,
	memberCloser MemberCloser,
) WorkspaceUseCase {
	return &workspaceInteractor{
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
		userNoteRepo:  userNoteRepo,
		permissionSvc: permissionSvc,
		recorder:      recorder,
		memberCloser:  memberCloser,
	}
}

func (i *workspaceInteractor) GetWorkspacesByUserID(ctx context.Context, userID string) (*GetWorkspacesOutput, error) {
	workspaces, err := i.workspaceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspaces: %w", err)
	}

	output := &GetWorkspacesOutput{
		Workspaces: make([]WorkspaceOutput, 0, len(workspaces)),
	}

	for _, ws := range workspaces {
		member, err := i.workspaceRepo.FindMember(ctx, ws.ID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get member info: %w", err)
		}
		if member == nil {
			continue
		}

		output.Workspaces = append(output.Workspaces, newWorkspaceOutput(ws, member.Role))
	}

	return output, nil
}

func (i *workspaceInteractor) GetWorkspace(ctx context.Context, input GetWorkspaceInput) (*GetWorkspaceOutput, error) {
	member, err := i.workspaceRepo.FindMember(ctx, input.ID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	ws, err := i.workspaceRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFound
	}

	return &GetWorkspaceOutput{
		Workspace: newWorkspaceOutput(ws, member.Role),
	}, nil
}

func (i *workspaceInteractor) CreateWorkspace(ctx context.Context, input CreateWorkspaceInput) (*CreateWorkspaceOutput, error) {
	// Validate slug
	if err := entity.ValidateWorkspaceSlug(input.ID); err != nil {
		return nil, err
	}

	// Check duplication
	exists, err := i.workspaceRepo.ExistsByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check workspace id: %w", err)
	}
	if exists {
		return nil, errors.New("このワークスペースIDは既に使用されています")
	}

	workspace := &entity.Workspace{
		ID:          input.ID,
		Name:        input.Name,
		Description: input.Description,
		IconURL:     input.IconURL,
		IsPublic:    input.IsPublic,
		CreatedBy:   input.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := i.workspaceRepo.Create(ctx, workspace); err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	member := &entity.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      input.CreatedBy,
		Role:        entity.WorkspaceRoleOwner,
		JoinedAt:    time.Now(),
	}

	if err := i.workspaceRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add creator as owner: %w", err)
	}

	return &CreateWorkspaceOutput{
		Workspace: newWorkspaceOutput(workspace, entity.WorkspaceRoleOwner),
	}, nil
}

func (i *workspaceInteractor) UpdateWorkspace(ctx context.Context, input UpdateWorkspaceInput) (*UpdateWorkspaceOutput, error) {
	member, err := i.workspaceRepo.FindMember(ctx, input.ID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if member == nil || (member.Role != entity.WorkspaceRoleOwner && member.Role != entity.WorkspaceRoleAdmin) {
		return nil, ErrUnauthorized
	}

	ws, err := i.workspaceRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFound
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
	ws.UpdatedAt = time.Now()

	if err := i.workspaceRepo.Update(ctx, ws); err != nil {
		return nil, fmt.Errorf("failed to update workspace: %w", err)
	}

	return &UpdateWorkspaceOutput{
		Workspace: newWorkspaceOutput(ws, member.Role),
	}, nil
}

func (i *workspaceInteractor) DeleteWorkspace(ctx context.Context, input DeleteWorkspaceInput) (*DeleteWorkspaceOutput, error) {
	member, err := i.workspaceRepo.FindMember(ctx, input.ID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if member == nil || member.Role != entity.WorkspaceRoleOwner {
		return nil, ErrUnauthorized
	}

	if err := i.workspaceRepo.Delete(ctx, input.ID); err != nil {
		return nil, fmt.Errorf("failed to delete workspace: %w", err)
	}

	return &DeleteWorkspaceOutput{Success: true}, nil
}

func (i *workspaceInteractor) ListMembers(ctx context.Context, input ListMembersInput) (*ListMembersOutput, error) {
	member, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	members, err := i.workspaceRepo.FindMembersByWorkspaceID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}

	// ユーザーIDを収集
	userIDs := make([]string, 0, len(members))
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}

	// ユーザー情報を一括取得
	users, err := i.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// ユーザー情報をマップに格納
	userMap := make(map[string]*entity.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	nicknames, err := i.userNoteRepo.FindNicknames(ctx, input.RequesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get nicknames: %w", err)
	}

	output := &ListMembersOutput{Members: make([]MemberInfo, 0, len(members))}
	for _, m := range members {
		user := userMap[m.UserID]
		memberInfo := MemberInfo{
			UserID:      m.UserID,
			Role:        string(m.Role),
			JoinedAt:    m.JoinedAt,
			SuspendedAt: m.SuspendedAt,
		}
		if nickname, ok := nicknames[m.UserID]; ok {
			memberInfo.Nickname = &nickname
		}
		if user != nil {
			memberInfo.Email = user.Email
			memberInfo.DisplayName = user.DisplayName
			memberInfo.AvatarURL = user.AvatarURL
			memberInfo.Bio = user.Bio
			memberInfo.Timezone = user.Preferences.Timezone
			memberInfo.Links = user.Links
		}
		output.Members = append(output.Members, memberInfo)
	}

	return output, nil
}

func (i *workspaceInteractor) UpdateMemberRole(ctx context.Context, input UpdateMemberRoleInput) (*MemberActionOutput, error) {
	requester, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UpdaterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check requester membership: %w", err)
	}
	if requester == nil || (requester.Role != entity.WorkspaceRoleOwner && requester.Role != entity.WorkspaceRoleAdmin) {
		return nil, ErrUnauthorized
	}

	if err := validateWorkspaceRole(input.Role); err != nil {
		return nil, err
	}

	target, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target member: %w", err)
	}
	if target == nil {
		return nil, ErrWorkspaceNotFound
	}

	// owner の降格と owner への昇格は owner 本人にのみ許可する
	isOwnerChange := target.Role == entity.WorkspaceRoleOwner || entity.WorkspaceRole(input.Role) == entity.WorkspaceRoleOwner
	if isOwnerChange && requester.Role != entity.WorkspaceRoleOwner {
		return nil, ErrCannotChangeOwnerRole
	}
	if input.UserID == input.UpdaterID {
		return nil, ErrCannotChangeOwnerRole
	}

	if err := i.workspaceRepo.UpdateMemberRole(ctx, input.WorkspaceID, input.UserID, entity.WorkspaceRole(input.Role)); err != nil {
		return nil, fmt.Errorf("failed to update member role: %w", err)
	}

	label := input.UserID
	if user, err := i.userRepo.FindByID(ctx, input.UserID); err == nil && user != nil {
		label = user.DisplayName
	}
	i.recorder.Record(ctx, entity.AuditLog{
		WorkspaceID: input.WorkspaceID,
		ActorID:     &input.UpdaterID,
		Action:      entity.AuditActionMemberRoleChanged,
		TargetType:  entity.AuditTargetUser,
		TargetID:    input.UserID,
		TargetLabel: label,
		Metadata:    map[string]string{"from": string(target.Role), "to": input.Role},
	})

	return &MemberActionOutput{Success: true}, nil
}

func (i *workspaceInteractor) RemoveMember(ctx context.Context, input RemoveMemberInput) (*MemberActionOutput, error) {
	requester, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.RemoverID)
	if err != nil {
		return nil, fmt.Errorf("failed to check requester membership: %w", err)
	}
	if requester == nil || (requester.Role != entity.WorkspaceRoleOwner && requester.Role != entity.WorkspaceRoleAdmin) {
		return nil, ErrUnauthorized
	}

	target, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target member: %w", err)
	}
	if target == nil {
		return nil, ErrWorkspaceNotFound
	}
	if target.Role == entity.WorkspaceRoleOwner {
		return nil, ErrCannotRemoveOwner
	}

	if err := i.workspaceRepo.RemoveMember(ctx, input.WorkspaceID, input.UserID); err != nil {
		return nil, fmt.Errorf("failed to remove member: %w", err)
	}
	i.memberCloser.CloseWorkspaceUser(input.WorkspaceID, input.UserID)

	return &MemberActionOutput{Success: true}, nil
}

func validateWorkspaceRole(role string) error {
	switch entity.WorkspaceRole(role) {
	case entity.WorkspaceRoleOwner, entity.WorkspaceRoleAdmin, entity.WorkspaceRoleMember, entity.WorkspaceRoleGuest:
		return nil
	default:
		return ErrInvalidRole
	}
}

// ListPublicWorkspaces returns public workspaces with member counts and joined flags
func (i *workspaceInteractor) ListPublicWorkspaces(ctx context.Context, userID string) (*ListPublicWorkspacesOutput, error) {
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

	output := &ListPublicWorkspacesOutput{Workspaces: make([]PublicWorkspaceItem, 0, len(workspaces))}
	for _, w := range workspaces {
		count, err := i.workspaceRepo.CountMembers(ctx, w.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to count members: %w", err)
		}
		output.Workspaces = append(output.Workspaces, PublicWorkspaceItem{
			ID:          w.ID,
			Name:        w.Name,
			Description: w.Description,
			IconURL:     w.IconURL,
			MemberCount: count,
			IsJoined:    joinedMap[w.ID],
			CreatedAt:   w.CreatedAt,
		})
	}
	return output, nil
}

// JoinPublicWorkspace joins a user to a public workspace.
func (i *workspaceInteractor) JoinPublicWorkspace(ctx context.Context, input JoinPublicWorkspaceInput) (*MemberActionOutput, error) {
	ws, err := i.workspaceRepo.FindByID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFound
	}
	if !ws.IsPublic && !ws.SignupEnabled {
		return nil, errors.New("このワークスペースは公開されていません")
	}

	existing, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if existing != nil {
		return nil, errors.New("既にこのワークスペースに参加しています")
	}

	member := &entity.WorkspaceMember{
		WorkspaceID: input.WorkspaceID,
		UserID:      input.UserID,
		Role:        entity.WorkspaceRoleMember,
		JoinedAt:    time.Now(),
	}
	if err := i.workspaceRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to join workspace: %w", err)
	}
	return &MemberActionOutput{Success: true}, nil
}

func newWorkspaceOutput(ws *entity.Workspace, role entity.WorkspaceRole) WorkspaceOutput {
	return WorkspaceOutput{
		ID:                 ws.ID,
		Name:               ws.Name,
		Description:        ws.Description,
		IconURL:            ws.IconURL,
		IsPublic:           ws.IsPublic,
		SignupEnabled:      ws.SignupEnabled,
		EmailSignupEnabled: ws.EmailSignupEnabled,
		Role:               string(role),
		CreatedBy:          ws.CreatedBy,
		CreatedAt:          ws.CreatedAt,
		UpdatedAt:          ws.UpdatedAt,
	}
}

// GetSignupInfo は参加リンクの画面に出す情報を返します。登録を許可していなければ存在しないものとして扱います
func (i *workspaceInteractor) GetSignupInfo(ctx context.Context, workspaceID string) (*SignupInfoOutput, error) {
	ws, err := i.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	if ws == nil || !ws.SignupEnabled {
		return nil, ErrWorkspaceNotFound
	}
	return &SignupInfoOutput{ID: ws.ID, Name: ws.Name, IconURL: ws.IconURL, EmailSignupEnabled: ws.EmailSignupEnabled}, nil
}
