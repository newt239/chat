package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/infrastructure/utils"
	workspaceuc "github.com/newt239/chat/internal/usecase/workspace"
)

type WorkspaceHandler struct {
	workspaceUC workspaceuc.WorkspaceUseCase
}

func NewWorkspaceHandler(workspaceUC workspaceuc.WorkspaceUseCase) *WorkspaceHandler {
	return &WorkspaceHandler{workspaceUC: workspaceUC}
}

// CreateWorkspaceRequest はワークスペース作成リクエストの構造体です
type CreateWorkspaceRequest struct {
	Name        string `json:"name" validate:"required,min=1"`
	Description string `json:"description"`
}

// UpdateWorkspaceRequest はワークスペース更新リクエストの構造体です
type UpdateWorkspaceRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1"`
	Description *string `json:"description,omitempty"`
}

// AddMemberRequest はメンバー追加リクエストの構造体です
type AddMemberRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Role   string `json:"role" validate:"required,oneof=owner admin member"`
}

// UpdateMemberRoleRequest はメンバーロール更新リクエストの構造体です
type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=owner admin member"`
}

// GetWorkspaces はワークスペース一覧を取得します
func (h *WorkspaceHandler) GetWorkspaces(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	workspaces, err := h.workspaceUC.GetWorkspacesByUserID(c.Request().Context(), userID)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, workspaces)
}

// CreateWorkspace はワークスペースを作成します
func (h *WorkspaceHandler) CreateWorkspace(c echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	var req CreateWorkspaceRequest
	if err := utils.ValidateRequest(c, &req); err != nil {
		return err
	}

	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	input := workspaceuc.CreateWorkspaceInput{
		Name:        req.Name,
		Description: description,
		CreatedBy:   userID,
	}

	workspace, err := h.workspaceUC.CreateWorkspace(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusCreated, workspace)
}

// GetWorkspace はワークスペース詳細を取得します
func (h *WorkspaceHandler) GetWorkspace(c echo.Context) error {
	workspaceID, err := utils.GetParamFromContext(c, "id")
	if err != nil {
		return err
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	input := workspaceuc.GetWorkspaceInput{
		ID:     workspaceID,
		UserID: userID,
	}

	workspace, err := h.workspaceUC.GetWorkspace(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, workspace)
}

// UpdateWorkspace はワークスペースを更新します
func (h *WorkspaceHandler) UpdateWorkspace(c echo.Context) error {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace ID is required")
	}

	var req UpdateWorkspaceRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := workspaceuc.UpdateWorkspaceInput{
		ID:          workspaceID,
		Name:        req.Name,
		Description: req.Description,
	}

	workspace, err := h.workspaceUC.UpdateWorkspace(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, workspace)
}

// DeleteWorkspace はワークスペースを削除します
func (h *WorkspaceHandler) DeleteWorkspace(c echo.Context) error {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	input := workspaceuc.DeleteWorkspaceInput{
		ID:     workspaceID,
		UserID: userID,
	}

	_, err := h.workspaceUC.DeleteWorkspace(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// ListMembers はワークスペースメンバー一覧を取得します
func (h *WorkspaceHandler) ListMembers(c echo.Context) error {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	input := workspaceuc.ListMembersInput{
		WorkspaceID: workspaceID,
		RequesterID: userID,
	}

	members, err := h.workspaceUC.ListMembers(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, members)
}

// AddMember はワークスペースにメンバーを追加します
func (h *WorkspaceHandler) AddMember(c echo.Context) error {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace ID is required")
	}

	var req AddMemberRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := workspaceuc.AddMemberInput{
		WorkspaceID: workspaceID,
		UserID:      req.UserID,
		Role:        req.Role,
	}

	member, err := h.workspaceUC.AddMember(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusCreated, member)
}

// UpdateMemberRole はメンバーのロールを更新します
func (h *WorkspaceHandler) UpdateMemberRole(c echo.Context) error {
	workspaceID := c.Param("id")
	userID := c.Param("userId")
	if workspaceID == "" || userID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace ID and User ID are required")
	}

	var req UpdateMemberRoleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := workspaceuc.UpdateMemberRoleInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        req.Role,
	}

	member, err := h.workspaceUC.UpdateMemberRole(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, member)
}

// RemoveMember はワークスペースからメンバーを削除します
func (h *WorkspaceHandler) RemoveMember(c echo.Context) error {
	workspaceID := c.Param("id")
	userID := c.Param("userId")
	if workspaceID == "" || userID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace ID and User ID are required")
	}

	removerID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	input := workspaceuc.RemoveMemberInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		RemoverID:   removerID,
	}

	_, err := h.workspaceUC.RemoveMember(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
