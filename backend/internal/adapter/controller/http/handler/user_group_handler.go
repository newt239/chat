package handler

import (
	"net/http"

	usergroupuc "github.com/example/chat/internal/usecase/user_group"
	"github.com/labstack/echo/v4"
)

type UserGroupHandler struct {
	userGroupUC usergroupuc.UserGroupUseCase
}

func NewUserGroupHandler(userGroupUC usergroupuc.UserGroupUseCase) *UserGroupHandler {
	return &UserGroupHandler{userGroupUC: userGroupUC}
}

// CreateUserGroupRequest はユーザーグループ作成リクエストの構造体です
type CreateUserGroupRequest struct {
	Name        string `json:"name" validate:"required,min=1"`
	Description string `json:"description"`
	WorkspaceID string `json:"workspace_id" validate:"required"`
}

// UpdateUserGroupRequest はユーザーグループ更新リクエストの構造体です
type UpdateUserGroupRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1"`
	Description *string `json:"description,omitempty"`
}

// AddUserGroupMemberRequest はユーザーグループメンバー追加リクエストの構造体です
type AddUserGroupMemberRequest struct {
	UserID string `json:"user_id" validate:"required"`
}

// CreateUserGroup はユーザーグループを作成します
func (h *UserGroupHandler) CreateUserGroup(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	var req CreateUserGroupRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	input := usergroupuc.CreateUserGroupInput{
		Name:        req.Name,
		Description: description,
		WorkspaceID: req.WorkspaceID,
		CreatedBy:   userID,
	}

	userGroup, err := h.userGroupUC.CreateUserGroup(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusCreated, userGroup)
}

// ListUserGroups はユーザーグループ一覧を取得します
func (h *UserGroupHandler) ListUserGroups(c echo.Context) error {
	workspaceID := c.QueryParam("workspace_id")
	if workspaceID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	input := usergroupuc.ListUserGroupsInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}

	userGroups, err := h.userGroupUC.ListUserGroups(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, userGroups)
}

// GetUserGroup はユーザーグループ詳細を取得します
func (h *UserGroupHandler) GetUserGroup(c echo.Context) error {
	userGroupID := c.Param("id")
	if userGroupID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "User Group ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	input := usergroupuc.GetUserGroupInput{
		ID:     userGroupID,
		UserID: userID,
	}

	userGroup, err := h.userGroupUC.GetUserGroup(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, userGroup)
}

// UpdateUserGroup はユーザーグループを更新します
func (h *UserGroupHandler) UpdateUserGroup(c echo.Context) error {
	userGroupID := c.Param("id")
	if userGroupID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "User Group ID is required")
	}

	var req UpdateUserGroupRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := usergroupuc.UpdateUserGroupInput{
		ID:          userGroupID,
		Name:        req.Name,
		Description: req.Description,
	}

	userGroup, err := h.userGroupUC.UpdateUserGroup(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, userGroup)
}

// DeleteUserGroup はユーザーグループを削除します
func (h *UserGroupHandler) DeleteUserGroup(c echo.Context) error {
	userGroupID := c.Param("id")
	if userGroupID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "User Group ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	input := usergroupuc.DeleteUserGroupInput{
		ID:        userGroupID,
		DeletedBy: userID,
	}

	_, err := h.userGroupUC.DeleteUserGroup(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// AddMember はユーザーグループにメンバーを追加します
func (h *UserGroupHandler) AddMember(c echo.Context) error {
	userGroupID := c.Param("id")
	if userGroupID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "User Group ID is required")
	}

	var req AddUserGroupMemberRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	addedBy, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	input := usergroupuc.AddMemberInput{
		GroupID: userGroupID,
		UserID:  req.UserID,
		AddedBy: addedBy,
	}

	member, err := h.userGroupUC.AddMember(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusCreated, member)
}

// RemoveMember はユーザーグループからメンバーを削除します
func (h *UserGroupHandler) RemoveMember(c echo.Context) error {
	userGroupID := c.Param("id")
	userID := c.Param("userId")
	if userGroupID == "" || userID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "User Group ID and User ID are required")
	}

	removedBy, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	input := usergroupuc.RemoveMemberInput{
		GroupID:   userGroupID,
		UserID:    userID,
		RemovedBy: removedBy,
	}

	_, err := h.userGroupUC.RemoveMember(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// ListMembers はユーザーグループメンバー一覧を取得します
func (h *UserGroupHandler) ListMembers(c echo.Context) error {
	userGroupID := c.Param("id")
	if userGroupID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "User Group ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	input := usergroupuc.ListMembersInput{
		GroupID: userGroupID,
		UserID:  userID,
	}

	members, err := h.userGroupUC.ListMembers(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, members)
}
