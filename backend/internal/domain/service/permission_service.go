package service

import (
	"context"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domainerrors "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

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

func NewPermissionService(
	workspaceRepo domainrepository.WorkspaceRepository,
	permissionRepo domainrepository.PermissionRepository,
) PermissionService {
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
	member, err := s.workspaceRepo.FindMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, domainerrors.ErrUnauthorized
	}
	if member.Role == entity.WorkspaceRoleOwner {
		return member, nil
	}

	matrix, err := s.Matrix(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if !matrix.Allows(member.Role, permission) {
		return nil, domainerrors.ErrUnauthorized
	}
	return member, nil
}
