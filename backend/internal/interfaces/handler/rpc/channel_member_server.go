package rpc

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	channelmemberuc "github.com/newt239/chat/internal/usecase/channelmember"
)

type ChannelMemberServer struct {
	UC *channelmemberuc.Interactor
}

func (s *ChannelMemberServer) ListChannelMembers(ctx context.Context, req *chatv1.ListChannelMembersRequest) (*chatv1.ListChannelMembersResponse, error) {
	out, err := s.UC.ListMembers(ctx, req.ChannelId, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListChannelMembersResponse{Members: presenter.ConvertAll(out, presenter.ChannelMember)}, nil
}

func (s *ChannelMemberServer) InviteChannelMember(ctx context.Context, req *chatv1.InviteChannelMemberRequest) (*chatv1.InviteChannelMemberResponse, error) {
	role := entity.ChannelRoleMember
	if req.Role != chatv1.ChannelRole_CHANNEL_ROLE_UNSPECIFIED {
		role = fromProto(presenter.ChannelRoles, req.Role)
	}
	input := channelmemberuc.MemberInput{ChannelID: req.ChannelId, OperatorID: userIDFrom(ctx), TargetUserID: req.UserId, Role: role}
	return &chatv1.InviteChannelMemberResponse{}, s.UC.InviteMember(ctx, input)
}

func (s *ChannelMemberServer) JoinChannel(ctx context.Context, req *chatv1.JoinChannelRequest) (*chatv1.JoinChannelResponse, error) {
	return &chatv1.JoinChannelResponse{}, s.UC.JoinPublicChannel(ctx, req.ChannelId, userIDFrom(ctx))
}

func (s *ChannelMemberServer) LeaveChannel(ctx context.Context, req *chatv1.LeaveChannelRequest) (*chatv1.LeaveChannelResponse, error) {
	return &chatv1.LeaveChannelResponse{}, s.UC.LeaveChannel(ctx, req.ChannelId, userIDFrom(ctx))
}

func (s *ChannelMemberServer) RemoveChannelMember(ctx context.Context, req *chatv1.RemoveChannelMemberRequest) (*chatv1.RemoveChannelMemberResponse, error) {
	input := channelmemberuc.MemberInput{ChannelID: req.ChannelId, OperatorID: userIDFrom(ctx), TargetUserID: req.UserId}
	return &chatv1.RemoveChannelMemberResponse{}, s.UC.RemoveMember(ctx, input)
}

func (s *ChannelMemberServer) UpdateChannelMemberRole(ctx context.Context, req *chatv1.UpdateChannelMemberRoleRequest) (*chatv1.UpdateChannelMemberRoleResponse, error) {
	input := channelmemberuc.MemberInput{ChannelID: req.ChannelId, OperatorID: userIDFrom(ctx), TargetUserID: req.UserId, Role: fromProto(presenter.ChannelRoles, req.Role)}
	return &chatv1.UpdateChannelMemberRoleResponse{}, s.UC.UpdateMemberRole(ctx, input)
}
