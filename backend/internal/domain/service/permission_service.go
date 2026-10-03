package service

import (
	"context"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

// EnsureMember は停止されずにワークスペースに参加しているメンバーを返します
func EnsureMember(ctx context.Context, workspaceRepo domainrepository.WorkspaceRepository, workspaceID, userID string) (*entity.WorkspaceMember, error) {
	member, err := workspaceRepo.FindMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, domerr.ErrUnauthorized
	}
	return member, nil
}

// EnsureAdmin は owner / admin のメンバーを返します
func EnsureAdmin(ctx context.Context, workspaceRepo domainrepository.WorkspaceRepository, workspaceID, userID string) (*entity.WorkspaceMember, error) {
	member, err := EnsureMember(ctx, workspaceRepo, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if !member.IsAdmin() {
		return nil, domerr.ErrUnauthorized
	}
	return member, nil
}

// PermissionService はワークスペースの権限設定に基づく操作の可否を判定します
type PermissionService interface {
	Matrix(ctx context.Context, workspaceID string) (entity.PermissionMatrix, error)
	// Ensure は停止中でないメンバーで、操作が許可されている場合にそのメンバーを返します
	Ensure(ctx context.Context, workspaceID, userID string, permission entity.Permission) (*entity.WorkspaceMember, error)
}

type permissionService struct {
	workspaceRepo  domainrepository.WorkspaceRepository
	permissionRepo domainrepository.PermissionRepository
}

func NewPermissionService(workspaceRepo domainrepository.WorkspaceRepository, permissionRepo domainrepository.PermissionRepository) PermissionService {
	return &permissionService{workspaceRepo: workspaceRepo, permissionRepo: permissionRepo}
}

func (s *permissionService) Matrix(ctx context.Context, workspaceID string) (entity.PermissionMatrix, error) {
	overrides, err := s.permissionRepo.FindOverrides(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load permissions: %w", err)
	}
	return entity.DefaultPermissionMatrix().Apply(overrides), nil
}

func (s *permissionService) Ensure(ctx context.Context, workspaceID, userID string, permission entity.Permission) (*entity.WorkspaceMember, error) {
	member, err := EnsureMember(ctx, s.workspaceRepo, workspaceID, userID)
	if err != nil || member.Role == entity.WorkspaceRoleOwner {
		return member, err
	}
	matrix, err := s.Matrix(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if !matrix.Allows(member.Role, permission) {
		return nil, domerr.ErrUnauthorized
	}
	return member, nil
}
