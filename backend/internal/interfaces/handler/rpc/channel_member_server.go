package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	channelmemberuc "github.com/newt239/chat/internal/usecase/channelmember"
)

type ChannelMemberServer struct {
	UC channelmemberuc.ChannelMemberUseCase
}

func (s *ChannelMemberServer) ListChannelMembers(ctx context.Context, req *chatv1.ListChannelMembersRequest) (*chatv1.ListChannelMembersResponse, error) {
	out, err := s.UC.ListMembers(ctx, channelmemberuc.ListMembersInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.ListChannelMembersResponse{Members: presenter.ConvertAll(out.Members, presenter.ChannelMember)}, nil
}

func (s *ChannelMemberServer) InviteChannelMember(ctx context.Context, req *chatv1.InviteChannelMemberRequest) (*chatv1.InviteChannelMemberResponse, error) {
	role := chatv1.ChannelRole_CHANNEL_ROLE_MEMBER
	if req.Role != chatv1.ChannelRole_CHANNEL_ROLE_UNSPECIFIED {
		role = req.Role
	}
	err := s.UC.InviteMember(ctx, channelmemberuc.InviteMemberInput{
		ChannelID:    req.ChannelId,
		OperatorID:   userIDFrom(ctx),
		TargetUserID: req.UserId,
		Role:         presenter.ChannelRoleFromProto(role),
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.InviteChannelMemberResponse{}, nil
}

func (s *ChannelMemberServer) JoinChannel(ctx context.Context, req *chatv1.JoinChannelRequest) (*chatv1.JoinChannelResponse, error) {
	if err := s.UC.JoinPublicChannel(ctx, channelmemberuc.JoinChannelInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx)}); err != nil {
		return nil, err
	}
	return &chatv1.JoinChannelResponse{}, nil
}

func (s *ChannelMemberServer) LeaveChannel(ctx context.Context, req *chatv1.LeaveChannelRequest) (*chatv1.LeaveChannelResponse, error) {
	if err := s.UC.LeaveChannel(ctx, channelmemberuc.LeaveChannelInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx)}); err != nil {
		return nil, err
	}
	return &chatv1.LeaveChannelResponse{}, nil
}

func (s *ChannelMemberServer) RemoveChannelMember(ctx context.Context, req *chatv1.RemoveChannelMemberRequest) (*chatv1.RemoveChannelMemberResponse, error) {
	err := s.UC.RemoveMember(ctx, channelmemberuc.RemoveMemberInput{ChannelID: req.ChannelId, OperatorID: userIDFrom(ctx), TargetUserID: req.UserId})
	if err != nil {
		return nil, err
	}
	return &chatv1.RemoveChannelMemberResponse{}, nil
}

func (s *ChannelMemberServer) UpdateChannelMemberRole(ctx context.Context, req *chatv1.UpdateChannelMemberRoleRequest) (*chatv1.UpdateChannelMemberRoleResponse, error) {
	err := s.UC.UpdateMemberRole(ctx, channelmemberuc.UpdateMemberRoleInput{
		ChannelID:    req.ChannelId,
		OperatorID:   userIDFrom(ctx),
		TargetUserID: req.UserId,
		Role:         presenter.ChannelRoleFromProto(req.Role),
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdateChannelMemberRoleResponse{}, nil
}
