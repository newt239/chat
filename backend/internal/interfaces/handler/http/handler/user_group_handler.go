package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/newt239/chat/internal/openapi_gen"
	usergroupuc "github.com/newt239/chat/internal/usecase/user_group"
)

type UserGroupHandler struct {
	UserGroupUC usergroupuc.UserGroupUseCase
}

// AddUserGroupMemberRequest はユーザーグループメンバー追加リクエストの構造体です
type AddUserGroupMemberRequest struct {
	UserID string `json:"user_id" validate:"required"`
}

// CreateUserGroup はユーザーグループを作成します
func (h *UserGroupHandler) CreateUserGroup(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "コンテキストにユーザーIDが含まれていません")
	}

	var req openapi.CreateUserGroupRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "リクエストボディが不正です")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := usergroupuc.CreateUserGroupInput{
		Name:        req.Name,
		Description: req.Description,
		WorkspaceID: req.WorkspaceId.String(),
		CreatedBy:   userID,
	}

	userGroup, err := h.UserGroupUC.CreateUserGroup(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusCreated, userGroup)
}

// ListUserGroups implements ServerInterface.ListUserGroups
func (h *UserGroupHandler) ListUserGroups(ctx echo.Context, params openapi.ListUserGroupsParams) error {
	userID, ok := ctx.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "コンテキストにユーザーIDが含まれていません")
	}

	input := usergroupuc.ListUserGroupsInput{
		WorkspaceID: params.WorkspaceId,
		UserID:      userID,
	}

	userGroups, err := h.UserGroupUC.ListUserGroups(ctx.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return ctx.JSON(http.StatusOK, userGroups)
}

// DeleteUserGroup implements ServerInterface.DeleteUserGroup
func (h *UserGroupHandler) DeleteUserGroup(ctx echo.Context, id openapi_types.UUID) error {
	userID, ok := ctx.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "コンテキストにユーザーIDが含まれていません")
	}

	input := usergroupuc.DeleteUserGroupInput{
		ID:        id.String(),
		DeletedBy: userID,
	}

	_, err := h.UserGroupUC.DeleteUserGroup(ctx.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetUserGroup implements ServerInterface.GetUserGroup
func (h *UserGroupHandler) GetUserGroup(ctx echo.Context, id openapi_types.UUID) error {
	userID, ok := ctx.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "コンテキストにユーザーIDが含まれていません")
	}

	input := usergroupuc.GetUserGroupInput{
		ID:     id.String(),
		UserID: userID,
	}

	userGroup, err := h.UserGroupUC.GetUserGroup(ctx.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return ctx.JSON(http.StatusOK, userGroup)
}

// UpdateUserGroup implements ServerInterface.UpdateUserGroup
func (h *UserGroupHandler) UpdateUserGroup(ctx echo.Context, id openapi_types.UUID) error {
	var req openapi.UpdateUserGroupRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "リクエストボディが不正です")
	}

	if err := ctx.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userID, ok := ctx.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "コンテキストにユーザーIDが含まれていません")
	}

	var name *string
	if req.Name != "" {
		name = &req.Name
	}

	input := usergroupuc.UpdateUserGroupInput{
		ID:          id.String(),
		Name:        name,
		Description: req.Description,
		UpdatedBy:   userID,
	}

	userGroup, err := h.UserGroupUC.UpdateUserGroup(ctx.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return ctx.JSON(http.StatusOK, userGroup)
}

// RemoveUserGroupMember implements ServerInterface.RemoveUserGroupMember
func (h *UserGroupHandler) RemoveUserGroupMember(ctx echo.Context, id openapi_types.UUID, params openapi.RemoveUserGroupMemberParams) error {
	removedBy, ok := ctx.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "コンテキストにユーザーIDが含まれていません")
	}

	input := usergroupuc.RemoveMemberInput{
		GroupID:   id.String(),
		UserID:    params.UserId.String(),
		RemovedBy: removedBy,
	}

	_, err := h.UserGroupUC.RemoveMember(ctx.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// ListUserGroupMembers implements ServerInterface.ListUserGroupMembers
func (h *UserGroupHandler) ListUserGroupMembers(ctx echo.Context, id openapi_types.UUID) error {
	userID, ok := ctx.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "コンテキストにユーザーIDが含まれていません")
	}

	input := usergroupuc.ListMembersInput{
		GroupID: id.String(),
		UserID:  userID,
	}

	members, err := h.UserGroupUC.ListMembers(ctx.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return ctx.JSON(http.StatusOK, members)
}

// AddUserGroupMember implements ServerInterface.AddUserGroupMember
func (h *UserGroupHandler) AddUserGroupMember(ctx echo.Context, id openapi_types.UUID) error {
	var req AddUserGroupMemberRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "リクエストボディが不正です")
	}

	if err := ctx.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	addedBy, ok := ctx.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "コンテキストにユーザーIDが含まれていません")
	}

	input := usergroupuc.AddMemberInput{
		GroupID: id.String(),
		UserID:  req.UserID,
		AddedBy: addedBy,
	}

	member, err := h.UserGroupUC.AddMember(ctx.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return ctx.JSON(http.StatusCreated, member)
}
