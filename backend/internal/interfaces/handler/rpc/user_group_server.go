package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	usergroupuc "github.com/newt239/chat/internal/usecase/user_group"
)

type UserGroupServer struct {
	UC usergroupuc.UserGroupUseCase
}

func (s *UserGroupServer) CreateUserGroup(ctx context.Context, req *chatv1.CreateUserGroupRequest) (*chatv1.CreateUserGroupResponse, error) {
	out, err := s.UC.CreateUserGroup(ctx, usergroupuc.CreateUserGroupInput{
		WorkspaceID: req.WorkspaceId,
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   userIDFrom(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.CreateUserGroupResponse{UserGroup: presenter.UserGroup(out.UserGroup)}, nil
}

func (s *UserGroupServer) ListUserGroups(ctx context.Context, req *chatv1.ListUserGroupsRequest) (*chatv1.ListUserGroupsResponse, error) {
	out, err := s.UC.ListUserGroups(ctx, usergroupuc.ListUserGroupsInput{WorkspaceID: req.WorkspaceId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.ListUserGroupsResponse{UserGroups: presenter.ConvertAll(out.UserGroups, presenter.UserGroup)}, nil
}

func (s *UserGroupServer) GetUserGroup(ctx context.Context, req *chatv1.GetUserGroupRequest) (*chatv1.GetUserGroupResponse, error) {
	out, err := s.UC.GetUserGroup(ctx, usergroupuc.GetUserGroupInput{ID: req.GroupId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.GetUserGroupResponse{UserGroup: presenter.UserGroup(out.UserGroup)}, nil
}

func (s *UserGroupServer) UpdateUserGroup(ctx context.Context, req *chatv1.UpdateUserGroupRequest) (*chatv1.UpdateUserGroupResponse, error) {
	out, err := s.UC.UpdateUserGroup(ctx, usergroupuc.UpdateUserGroupInput{
		ID:          req.GroupId,
		Name:        req.Name,
		Description: req.Description,
		UpdatedBy:   userIDFrom(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdateUserGroupResponse{UserGroup: presenter.UserGroup(out.UserGroup)}, nil
}

func (s *UserGroupServer) DeleteUserGroup(ctx context.Context, req *chatv1.DeleteUserGroupRequest) (*chatv1.DeleteUserGroupResponse, error) {
	if _, err := s.UC.DeleteUserGroup(ctx, usergroupuc.DeleteUserGroupInput{ID: req.GroupId, DeletedBy: userIDFrom(ctx)}); err != nil {
		return nil, err
	}
	return &chatv1.DeleteUserGroupResponse{}, nil
}

func (s *UserGroupServer) ListUserGroupMembers(ctx context.Context, req *chatv1.ListUserGroupMembersRequest) (*chatv1.ListUserGroupMembersResponse, error) {
	out, err := s.UC.ListMembers(ctx, usergroupuc.ListMembersInput{GroupID: req.GroupId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.ListUserGroupMembersResponse{Members: presenter.ConvertAll(out.Members, presenter.UserGroupMember)}, nil
}

func (s *UserGroupServer) AddUserGroupMember(ctx context.Context, req *chatv1.AddUserGroupMemberRequest) (*chatv1.AddUserGroupMemberResponse, error) {
	if _, err := s.UC.AddMember(ctx, usergroupuc.AddMemberInput{GroupID: req.GroupId, UserID: req.UserId, AddedBy: userIDFrom(ctx)}); err != nil {
		return nil, err
	}
	return &chatv1.AddUserGroupMemberResponse{}, nil
}

func (s *UserGroupServer) RemoveUserGroupMember(ctx context.Context, req *chatv1.RemoveUserGroupMemberRequest) (*chatv1.RemoveUserGroupMemberResponse, error) {
	if _, err := s.UC.RemoveMember(ctx, usergroupuc.RemoveMemberInput{GroupID: req.GroupId, UserID: req.UserId, RemovedBy: userIDFrom(ctx)}); err != nil {
		return nil, err
	}
	return &chatv1.RemoveUserGroupMemberResponse{}, nil
}
