package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	workspaceuc "github.com/newt239/chat/internal/usecase/workspace"
)

type WorkspaceServer struct {
	UC *workspaceuc.Interactor
}

func (s *WorkspaceServer) ListWorkspaces(ctx context.Context, _ *chatv1.ListWorkspacesRequest) (*chatv1.ListWorkspacesResponse, error) {
	out, err := s.UC.ListWorkspaces(ctx, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListWorkspacesResponse{Workspaces: presenter.ConvertAll(out, presenter.Workspace)}, nil
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
	return &chatv1.CreateWorkspaceResponse{Workspace: presenter.Workspace(*out)}, nil
}

func (s *WorkspaceServer) GetWorkspace(ctx context.Context, req *chatv1.GetWorkspaceRequest) (*chatv1.GetWorkspaceResponse, error) {
	out, err := s.UC.GetWorkspace(ctx, req.WorkspaceId, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.GetWorkspaceResponse{Workspace: presenter.Workspace(*out)}, nil
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
	return &chatv1.UpdateWorkspaceResponse{Workspace: presenter.Workspace(*out)}, nil
}

func (s *WorkspaceServer) DeleteWorkspace(ctx context.Context, req *chatv1.DeleteWorkspaceRequest) (*chatv1.DeleteWorkspaceResponse, error) {
	return &chatv1.DeleteWorkspaceResponse{}, s.UC.DeleteWorkspace(ctx, req.WorkspaceId, userIDFrom(ctx))
}

func (s *WorkspaceServer) ListPublicWorkspaces(ctx context.Context, _ *chatv1.ListPublicWorkspacesRequest) (*chatv1.ListPublicWorkspacesResponse, error) {
	out, err := s.UC.ListPublicWorkspaces(ctx, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListPublicWorkspacesResponse{Workspaces: presenter.ConvertAll(out, presenter.PublicWorkspace)}, nil
}

func (s *WorkspaceServer) JoinPublicWorkspace(ctx context.Context, req *chatv1.JoinPublicWorkspaceRequest) (*chatv1.JoinPublicWorkspaceResponse, error) {
	return &chatv1.JoinPublicWorkspaceResponse{}, s.UC.JoinPublicWorkspace(ctx, req.WorkspaceId, userIDFrom(ctx))
}

func (s *WorkspaceServer) ListMembers(ctx context.Context, req *chatv1.ListMembersRequest) (*chatv1.ListMembersResponse, error) {
	out, err := s.UC.ListMembers(ctx, req.WorkspaceId, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListMembersResponse{Members: presenter.ConvertAll(out, presenter.WorkspaceMember)}, nil
}

func (s *WorkspaceServer) UpdateMemberRole(ctx context.Context, req *chatv1.UpdateMemberRoleRequest) (*chatv1.UpdateMemberRoleResponse, error) {
	input := workspaceuc.MemberInput{WorkspaceID: req.WorkspaceId, UserID: req.UserId, OperatorID: userIDFrom(ctx), Role: fromProto(presenter.WorkspaceRoles, req.Role)}
	return &chatv1.UpdateMemberRoleResponse{}, s.UC.UpdateMemberRole(ctx, input)
}

func (s *WorkspaceServer) RemoveMember(ctx context.Context, req *chatv1.RemoveMemberRequest) (*chatv1.RemoveMemberResponse, error) {
	input := workspaceuc.MemberInput{WorkspaceID: req.WorkspaceId, UserID: req.UserId, OperatorID: userIDFrom(ctx)}
	return &chatv1.RemoveMemberResponse{}, s.UC.RemoveMember(ctx, input)
}

func (s *WorkspaceServer) GetWorkspaceSignupInfo(ctx context.Context, req *chatv1.GetWorkspaceSignupInfoRequest) (*chatv1.GetWorkspaceSignupInfoResponse, error) {
	out, err := s.UC.GetSignupInfo(ctx, req.WorkspaceId)
	if err != nil {
		return nil, err
	}
	return &chatv1.GetWorkspaceSignupInfoResponse{Id: out.ID, Name: out.Name, IconUrl: out.IconURL, EmailSignupEnabled: out.EmailSignupEnabled}, nil
}
