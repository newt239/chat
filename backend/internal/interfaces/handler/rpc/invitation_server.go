package rpc

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	invitationuc "github.com/newt239/chat/internal/usecase/invitation"
)

type InvitationServer struct {
	UC *invitationuc.Interactor
}

func (s *InvitationServer) CreateInvitation(ctx context.Context, req *chatv1.CreateInvitationRequest) (*chatv1.CreateInvitationResponse, error) {
	role := chatv1.WorkspaceRole_WORKSPACE_ROLE_MEMBER
	if req.Role != chatv1.WorkspaceRole_WORKSPACE_ROLE_UNSPECIFIED {
		role = req.Role
	}
	out, err := s.UC.Create(ctx, invitationuc.CreateInput{
		WorkspaceID: req.WorkspaceId,
		Email:       req.Email,
		Role:        entity.WorkspaceRole(presenter.WorkspaceRoleName(role)),
		RequestedBy: userIDFrom(ctx),
	})
	if err != nil {
		return nil, err
	}
	res := &chatv1.CreateInvitationResponse{AddedDirectly: out.AddedDirectly, Token: out.Token}
	if out.Invitation != nil {
		res.Invitation = presenter.Invitation(*out.Invitation)
	}
	return res, nil
}

func (s *InvitationServer) ListInvitations(ctx context.Context, req *chatv1.ListInvitationsRequest) (*chatv1.ListInvitationsResponse, error) {
	out, err := s.UC.List(ctx, req.WorkspaceId, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListInvitationsResponse{Invitations: presenter.ConvertAll(out, presenter.Invitation)}, nil
}

func (s *InvitationServer) RevokeInvitation(ctx context.Context, req *chatv1.RevokeInvitationRequest) (*chatv1.RevokeInvitationResponse, error) {
	if err := s.UC.Revoke(ctx, req.WorkspaceId, req.InvitationId, userIDFrom(ctx)); err != nil {
		return nil, err
	}
	return &chatv1.RevokeInvitationResponse{}, nil
}

func (s *InvitationServer) GetInvitation(ctx context.Context, req *chatv1.GetInvitationRequest) (*chatv1.GetInvitationResponse, error) {
	out, err := s.UC.Preview(ctx, req.Token)
	if err != nil {
		return nil, err
	}
	return &chatv1.GetInvitationResponse{WorkspaceName: out.WorkspaceName, Email: out.Email}, nil
}
