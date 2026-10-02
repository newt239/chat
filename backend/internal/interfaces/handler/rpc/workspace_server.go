package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	workspaceuc "github.com/newt239/chat/internal/usecase/workspace"
)

type WorkspaceServer struct {
	UC workspaceuc.WorkspaceUseCase
}

func (s *WorkspaceServer) ListWorkspaces(ctx context.Context, _ *chatv1.ListWorkspacesRequest) (*chatv1.ListWorkspacesResponse, error) {
	out, err := s.UC.GetWorkspacesByUserID(ctx, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListWorkspacesResponse{Workspaces: presenter.ConvertAll(out.Workspaces, presenter.Workspace)}, nil
}

func (s *WorkspaceServer) CreateWorkspace(ctx context.Context, req *chatv1.CreateWorkspaceRequest) (*chatv1.CreateWorkspaceResponse, error) {
	out, err := s.UC.CreateWorkspace(ctx, workspaceuc.CreateWorkspaceInput{
		ID:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		IconURL:     req.IconUrl,
		IsPublic:    req.IsPublic,
		CreatedBy:   userIDFrom(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.CreateWorkspaceResponse{Workspace: presenter.Workspace(out.Workspace)}, nil
}

func (s *WorkspaceServer) GetWorkspace(ctx context.Context, req *chatv1.GetWorkspaceRequest) (*chatv1.GetWorkspaceResponse, error) {
	out, err := s.UC.GetWorkspace(ctx, workspaceuc.GetWorkspaceInput{ID: req.WorkspaceId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.GetWorkspaceResponse{Workspace: presenter.Workspace(out.Workspace)}, nil
}

func (s *WorkspaceServer) UpdateWorkspace(ctx context.Context, req *chatv1.UpdateWorkspaceRequest) (*chatv1.UpdateWorkspaceResponse, error) {
	out, err := s.UC.UpdateWorkspace(ctx, workspaceuc.UpdateWorkspaceInput{
		ID:                 req.WorkspaceId,
		Name:               req.Name,
		Description:        req.Description,
		IconURL:            req.IconUrl,
		IsPublic:           req.IsPublic,
		SignupEnabled:      req.SignupEnabled,
		EmailSignupEnabled: req.EmailSignupEnabled,
		UserID:             userIDFrom(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdateWorkspaceResponse{Workspace: presenter.Workspace(out.Workspace)}, nil
}

func (s *WorkspaceServer) DeleteWorkspace(ctx context.Context, req *chatv1.DeleteWorkspaceRequest) (*chatv1.DeleteWorkspaceResponse, error) {
	if _, err := s.UC.DeleteWorkspace(ctx, workspaceuc.DeleteWorkspaceInput{ID: req.WorkspaceId, UserID: userIDFrom(ctx)}); err != nil {
		return nil, err
	}
	return &chatv1.DeleteWorkspaceResponse{}, nil
}

func (s *WorkspaceServer) ListPublicWorkspaces(ctx context.Context, _ *chatv1.ListPublicWorkspacesRequest) (*chatv1.ListPublicWorkspacesResponse, error) {
	out, err := s.UC.ListPublicWorkspaces(ctx, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListPublicWorkspacesResponse{Workspaces: presenter.ConvertAll(out.Workspaces, presenter.PublicWorkspace)}, nil
}

func (s *WorkspaceServer) JoinPublicWorkspace(ctx context.Context, req *chatv1.JoinPublicWorkspaceRequest) (*chatv1.JoinPublicWorkspaceResponse, error) {
	if _, err := s.UC.JoinPublicWorkspace(ctx, workspaceuc.JoinPublicWorkspaceInput{WorkspaceID: req.WorkspaceId, UserID: userIDFrom(ctx)}); err != nil {
		return nil, err
	}
	return &chatv1.JoinPublicWorkspaceResponse{}, nil
}

func (s *WorkspaceServer) ListMembers(ctx context.Context, req *chatv1.ListMembersRequest) (*chatv1.ListMembersResponse, error) {
	out, err := s.UC.ListMembers(ctx, workspaceuc.ListMembersInput{WorkspaceID: req.WorkspaceId, RequesterID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.ListMembersResponse{Members: presenter.ConvertAll(out.Members, presenter.WorkspaceMember)}, nil
}

func (s *WorkspaceServer) UpdateMemberRole(ctx context.Context, req *chatv1.UpdateMemberRoleRequest) (*chatv1.UpdateMemberRoleResponse, error) {
	_, err := s.UC.UpdateMemberRole(ctx, workspaceuc.UpdateMemberRoleInput{
		WorkspaceID: req.WorkspaceId,
		UserID:      req.UserId,
		UpdaterID:   userIDFrom(ctx),
		Role:        presenter.WorkspaceRoleFromProto(req.Role),
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdateMemberRoleResponse{}, nil
}

func (s *WorkspaceServer) RemoveMember(ctx context.Context, req *chatv1.RemoveMemberRequest) (*chatv1.RemoveMemberResponse, error) {
	_, err := s.UC.RemoveMember(ctx, workspaceuc.RemoveMemberInput{WorkspaceID: req.WorkspaceId, UserID: req.UserId, RemoverID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.RemoveMemberResponse{}, nil
}

func (s *WorkspaceServer) GetWorkspaceSignupInfo(ctx context.Context, req *chatv1.GetWorkspaceSignupInfoRequest) (*chatv1.GetWorkspaceSignupInfoResponse, error) {
	out, err := s.UC.GetSignupInfo(ctx, req.WorkspaceId)
	if err != nil {
		return nil, err
	}
	return &chatv1.GetWorkspaceSignupInfoResponse{Id: out.ID, Name: out.Name, IconUrl: out.IconURL, EmailSignupEnabled: out.EmailSignupEnabled}, nil
}
