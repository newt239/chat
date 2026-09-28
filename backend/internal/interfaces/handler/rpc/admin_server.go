package rpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	adminuc "github.com/newt239/chat/internal/usecase/admin"
)

type AdminServer struct {
	UC *adminuc.Interactor
}

type auditLogFilterRequest interface {
	GetWorkspaceId() string
	GetActorId() string
	GetActions() []chatv1.AuditAction
	GetSince() *timestamppb.Timestamp
	GetUntil() *timestamppb.Timestamp
}

func auditLogQuery(ctx context.Context, req auditLogFilterRequest) adminuc.AuditLogQuery {
	q := adminuc.AuditLogQuery{
		WorkspaceID: req.GetWorkspaceId(),
		RequesterID: userIDFrom(ctx),
		Actions:     presenter.AuditActionNames(req.GetActions()),
	}
	if actorID := req.GetActorId(); actorID != "" {
		q.ActorID = &actorID
	}
	if since := req.GetSince(); since != nil {
		t := since.AsTime()
		q.Since = &t
	}
	if until := req.GetUntil(); until != nil {
		t := until.AsTime()
		q.Until = &t
	}
	return q
}

func (s *AdminServer) ListAuditLogs(ctx context.Context, req *chatv1.ListAuditLogsRequest) (*chatv1.ListAuditLogsResponse, error) {
	limit := int(req.Limit)
	if limit == 0 {
		limit = 50
	}
	out, err := s.UC.ListAuditLogs(ctx, adminuc.ListAuditLogsInput{AuditLogQuery: auditLogQuery(ctx, req), Limit: limit, PageToken: req.PageToken})
	if err != nil {
		return nil, err
	}
	return &chatv1.ListAuditLogsResponse{Logs: presenter.ConvertAll(out.Logs, presenter.AuditLog), NextPageToken: out.NextPageToken}, nil
}

func (s *AdminServer) ExportAuditLogs(ctx context.Context, req *chatv1.ExportAuditLogsRequest) (*chatv1.ExportAuditLogsResponse, error) {
	out, err := s.UC.ExportAuditLogs(ctx, auditLogQuery(ctx, req))
	if err != nil {
		return nil, err
	}
	return &chatv1.ExportAuditLogsResponse{Content: out.Content, FileName: out.FileName}, nil
}

func (s *AdminServer) ListAdminMembers(ctx context.Context, req *chatv1.ListAdminMembersRequest) (*chatv1.ListAdminMembersResponse, error) {
	out, err := s.UC.ListMembers(ctx, adminuc.WorkspaceInput{WorkspaceID: req.WorkspaceId, RequesterID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.ListAdminMembersResponse{Members: presenter.ConvertAll(out, presenter.AdminMember)}, nil
}

func (s *AdminServer) SuspendMember(ctx context.Context, req *chatv1.SuspendMemberRequest) (*chatv1.SuspendMemberResponse, error) {
	input := adminuc.MemberActionInput{WorkspaceID: req.WorkspaceId, TargetUserID: req.UserId, OperatorID: userIDFrom(ctx)}
	if err := s.UC.SuspendMember(ctx, input); err != nil {
		return nil, err
	}
	return &chatv1.SuspendMemberResponse{}, nil
}

func (s *AdminServer) ResumeMember(ctx context.Context, req *chatv1.ResumeMemberRequest) (*chatv1.ResumeMemberResponse, error) {
	input := adminuc.MemberActionInput{WorkspaceID: req.WorkspaceId, TargetUserID: req.UserId, OperatorID: userIDFrom(ctx)}
	if err := s.UC.ResumeMember(ctx, input); err != nil {
		return nil, err
	}
	return &chatv1.ResumeMemberResponse{}, nil
}

type PermissionServer struct {
	UC *adminuc.Interactor
}

func (s *PermissionServer) GetPermissions(ctx context.Context, req *chatv1.GetPermissionsRequest) (*chatv1.GetPermissionsResponse, error) {
	out, err := s.UC.GetPermissions(ctx, adminuc.WorkspaceInput{WorkspaceID: req.WorkspaceId, RequesterID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return presenter.Permissions(*out), nil
}

func (s *PermissionServer) UpdatePermission(ctx context.Context, req *chatv1.UpdatePermissionRequest) (*chatv1.UpdatePermissionResponse, error) {
	err := s.UC.UpdatePermission(ctx, adminuc.UpdatePermissionInput{
		WorkspaceID: req.WorkspaceId,
		OperatorID:  userIDFrom(ctx),
		Role:        entity.WorkspaceRole(presenter.WorkspaceRoleName(req.Role)),
		Permission:  presenter.PermissionName(req.Permission),
		Allowed:     req.Allowed,
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdatePermissionResponse{}, nil
}
