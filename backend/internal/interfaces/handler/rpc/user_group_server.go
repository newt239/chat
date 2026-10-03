package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	usergroupuc "github.com/newt239/chat/internal/usecase/usergroup"
)

type UserGroupServer struct {
	UC *usergroupuc.Interactor
}

func (s *UserGroupServer) CreateUserGroup(ctx context.Context, req *chatv1.CreateUserGroupRequest) (*chatv1.CreateUserGroupResponse, error) {
	out, err := s.UC.Create(ctx, usergroupuc.CreateInput{WorkspaceID: req.WorkspaceId, Name: req.Name, Description: req.Description, CreatedBy: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.CreateUserGroupResponse{UserGroup: presenter.UserGroup(out)}, nil
}

func (s *UserGroupServer) ListUserGroups(ctx context.Context, req *chatv1.ListUserGroupsRequest) (*chatv1.ListUserGroupsResponse, error) {
	out, err := s.UC.List(ctx, req.WorkspaceId, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListUserGroupsResponse{UserGroups: presenter.ConvertAll(out, presenter.UserGroup)}, nil
}

func (s *UserGroupServer) UpdateUserGroup(ctx context.Context, req *chatv1.UpdateUserGroupRequest) (*chatv1.UpdateUserGroupResponse, error) {
	out, err := s.UC.Update(ctx, usergroupuc.UpdateInput{ID: req.GroupId, Name: req.Name, Description: req.Description, UpdatedBy: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdateUserGroupResponse{UserGroup: presenter.UserGroup(out)}, nil
}

func (s *UserGroupServer) DeleteUserGroup(ctx context.Context, req *chatv1.DeleteUserGroupRequest) (*chatv1.DeleteUserGroupResponse, error) {
	return &chatv1.DeleteUserGroupResponse{}, s.UC.Delete(ctx, req.GroupId, userIDFrom(ctx))
}

func (s *UserGroupServer) ListUserGroupMembers(ctx context.Context, req *chatv1.ListUserGroupMembersRequest) (*chatv1.ListUserGroupMembersResponse, error) {
	out, err := s.UC.ListMembers(ctx, req.GroupId, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListUserGroupMembersResponse{Members: presenter.ConvertAll(out, presenter.UserGroupMember)}, nil
}

func (s *UserGroupServer) AddUserGroupMember(ctx context.Context, req *chatv1.AddUserGroupMemberRequest) (*chatv1.AddUserGroupMemberResponse, error) {
	return &chatv1.AddUserGroupMemberResponse{}, s.UC.AddMember(ctx, usergroupuc.MemberInput{GroupID: req.GroupId, UserID: req.UserId, OperatorID: userIDFrom(ctx)})
}

func (s *UserGroupServer) RemoveUserGroupMember(ctx context.Context, req *chatv1.RemoveUserGroupMemberRequest) (*chatv1.RemoveUserGroupMemberResponse, error) {
	return &chatv1.RemoveUserGroupMemberResponse{}, s.UC.RemoveMember(ctx, usergroupuc.MemberInput{GroupID: req.GroupId, UserID: req.UserId, OperatorID: userIDFrom(ctx)})
}
