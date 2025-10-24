package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/example/chat/internal/usecase/user_group"
)

type UserGroupHandler struct {
	userGroupUseCase user_group.UserGroupUseCase
}

func NewUserGroupHandler(userGroupUseCase user_group.UserGroupUseCase) *UserGroupHandler {
	return &UserGroupHandler{
		userGroupUseCase: userGroupUseCase,
	}
}

// CreateUserGroup グループ作成
func (h *UserGroupHandler) CreateUserGroup(c echo.Context) error {
	var input user_group.CreateUserGroupInput
	if err := c.Bind(&input); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	output, err := h.userGroupUseCase.CreateUserGroup(c.Request().Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, output)
}

// UpdateUserGroup グループ更新
func (h *UserGroupHandler) UpdateUserGroup(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "group ID is required"})
		return echo.NewHTTPError(http.StatusBadRequest, "group ID is required")
	}

	var input user_group.UpdateUserGroupInput
	if err := c.Bind(&input); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	input.ID = id

	output, err := h.userGroupUseCase.UpdateUserGroup(c.Request().Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, output)
}

// DeleteUserGroup グループ削除
func (h *UserGroupHandler) DeleteUserGroup(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "group ID is required"})
		return echo.NewHTTPError(http.StatusBadRequest, "group ID is required")
	}

	// リクエストからユーザーIDを取得（認証ミドルウェアで設定される想定）
	userID, err := requireUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}

	input := user_group.DeleteUserGroupInput{
		ID:        id,
		DeletedBy: userID,
	}

	output, err := h.userGroupUseCase.DeleteUserGroup(c.Request().Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, output)
}

// GetUserGroup グループ取得
func (h *UserGroupHandler) GetUserGroup(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "group ID is required"})
		return echo.NewHTTPError(http.StatusBadRequest, "group ID is required")
	}

	// リクエストからユーザーIDを取得
	userID, err := requireUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}

	input := user_group.GetUserGroupInput{
		ID:     id,
		UserID: userID,
	}

	output, err := h.userGroupUseCase.GetUserGroup(c.Request().Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, output)
}

// ListUserGroups グループ一覧取得
func (h *UserGroupHandler) ListUserGroups(c echo.Context) error {
	workspaceID := c.QueryParam("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "workspace_id is required"})
		return echo.NewHTTPError(http.StatusBadRequest, "workspace_id is required")
	}

	// リクエストからユーザーIDを取得
	userID, err := requireUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, map[string]string{"error": "user not authenticated"})
		return echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}

	input := user_group.ListUserGroupsInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}

	output, err := h.userGroupUseCase.ListUserGroups(c.Request().Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, output)
}

// AddMember メンバー追加
func (h *UserGroupHandler) AddMember(c echo.Context) error {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "group ID is required"})
		return echo.NewHTTPError(http.StatusBadRequest, "group ID is required")
	}

	var requestBody struct {
		UserID string `json:"userId" binding:"required"`
	}
	if err := c.Bind(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// リクエストからユーザーIDを取得
	userID, err := requireUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}

	input := user_group.AddMemberInput{
		GroupID: groupID,
		UserID:  requestBody.UserID,
		AddedBy: userID,
	}

	output, err := h.userGroupUseCase.AddMember(c.Request().Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, output)
}

// RemoveMember メンバー削除
func (h *UserGroupHandler) RemoveMember(c echo.Context) error {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "group ID is required"})
		return echo.NewHTTPError(http.StatusBadRequest, "group ID is required")
	}

	userID := c.QueryParam("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "user_id is required"})
		return echo.NewHTTPError(http.StatusBadRequest, "user_id is required")
	}

	// リクエストからユーザーIDを取得
	requestUserID, err := requireUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}

	input := user_group.RemoveMemberInput{
		GroupID:   groupID,
		UserID:    userID,
		RemovedBy: requestUserID,
	}

	output, err := h.userGroupUseCase.RemoveMember(c.Request().Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, output)
}

// ListMembers メンバー一覧取得
func (h *UserGroupHandler) ListMembers(c echo.Context) error {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "group ID is required"})
		return echo.NewHTTPError(http.StatusBadRequest, "group ID is required")
	}

	// リクエストからユーザーIDを取得
	userID, err := requireUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, map[string]string{"error": "user not authenticated"})
		return echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}

	input := user_group.ListMembersInput{
		GroupID: groupID,
		UserID:  userID,
	}

	output, err := h.userGroupUseCase.ListMembers(c.Request().Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, output)
}
