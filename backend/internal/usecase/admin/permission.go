package admin

import (
	"context"
	"fmt"
	"strconv"

	"github.com/newt239/chat/internal/domain/entity"
	domainservice "github.com/newt239/chat/internal/domain/service"
)

// GetPermissions は権限の設定を返します。ボタンの出し分けに使うためメンバー全員が参照できます
func (i *Interactor) GetPermissions(ctx context.Context, input WorkspaceInput) (*PermissionsOutput, error) {
	member, err := domainservice.EnsureMember(ctx, i.workspaceRepo, input.WorkspaceID, input.RequesterID)
	if err != nil {
		return nil, err
	}

	matrix, err := i.permissionSvc.Matrix(ctx, input.WorkspaceID)
	if err != nil {
		return nil, err
	}
	return &PermissionsOutput{Matrix: matrix, RequesterRole: member.Role}, nil
}

func (i *Interactor) UpdatePermission(ctx context.Context, input UpdatePermissionInput) error {
	operator, err := domainservice.EnsureAdmin(ctx, i.workspaceRepo, input.WorkspaceID, input.OperatorID)
	if err != nil {
		return err
	}
	if !input.Permission.IsValid() || !entity.IsConfigurableRole(input.Role) {
		return ErrInvalidPermission
	}
	// 管理者が自分たちの権限を広げられないよう、管理者の列はオーナーだけが変更できる
	if input.Role == entity.WorkspaceRoleAdmin && operator.Role != entity.WorkspaceRoleOwner {
		return ErrOwnerOnlyPermissions
	}

	matrix, err := i.permissionSvc.Matrix(ctx, input.WorkspaceID)
	if err != nil {
		return err
	}
	if matrix.Allows(input.Role, input.Permission) == input.Allowed {
		return nil
	}

	override := entity.PermissionOverride{Role: input.Role, Permission: input.Permission, Allowed: input.Allowed}
	if err := i.permissionRepo.Upsert(ctx, input.WorkspaceID, override); err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}

	i.recorder.Record(ctx, entity.AuditLog{
		WorkspaceID: input.WorkspaceID,
		ActorID:     &input.OperatorID,
		Action:      entity.AuditActionPermissionChanged,
		TargetType:  entity.AuditTargetRole,
		TargetID:    string(input.Role),
		TargetLabel: string(input.Role),
		Metadata: map[string]string{
			"permission": string(input.Permission),
			"allowed":    strconv.FormatBool(input.Allowed),
		},
	})
	return nil
}
