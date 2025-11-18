package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/newt239/chat/internal/infrastructure/utils"
	"github.com/newt239/chat/internal/openapi_gen"
	workspaceuc "github.com/newt239/chat/internal/usecase/workspace"
)

type WorkspaceHandler struct {
	workspaceUC workspaceuc.WorkspaceUseCase
}

func NewWorkspaceHandler(workspaceUC workspaceuc.WorkspaceUseCase) *WorkspaceHandler {
	return &WorkspaceHandler{workspaceUC: workspaceUC}
}

// AddMemberRequest はメンバー追加リクエストの構造体です
// 注意: これはUserIDベースの追加用で、OpenAPIスキーマには定義がないため独自型を使用
type AddMemberRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Role   string `json:"role" validate:"required,oneof=owner admin member"`
}

// ListWorkspaces implements ServerInterface.ListWorkspaces
func (h *WorkspaceHandler) ListWorkspaces(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
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

	var req openapi.CreateWorkspaceRequest
	if err := utils.ValidateRequest(c, &req); err != nil {
		return err
	}

	var isPublic bool
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	input := workspaceuc.CreateWorkspaceInput{
		ID:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		IconURL:     req.IconUrl,
		IsPublic:    isPublic,
		CreatedBy:   userID,
	}

	workspace, err := h.workspaceUC.CreateWorkspace(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusCreated, workspace)
}

// GetWorkspace implements ServerInterface.GetWorkspace
func (h *WorkspaceHandler) GetWorkspace(c echo.Context, id string) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	input := workspaceuc.GetWorkspaceInput{
		ID:     id,
		UserID: userID,
	}

	workspace, err := h.workspaceUC.GetWorkspace(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, workspace)
}

// UpdateWorkspace implements ServerInterface.UpdateWorkspace
func (h *WorkspaceHandler) UpdateWorkspace(c echo.Context, id string) error {
	var req openapi.UpdateWorkspaceRequest
	if err := c.Bind(&req); err != nil {
		return utils.HandleBindError(err)
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	input := workspaceuc.UpdateWorkspaceInput{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		IconURL:     req.IconUrl,
		IsPublic:    req.IsPublic,
		UserID:      userID,
	}

	workspace, err := h.workspaceUC.UpdateWorkspace(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, workspace)
}

// ListPublicWorkspaces は公開ワークスペース一覧を返します
func (h *WorkspaceHandler) ListPublicWorkspaces(c echo.Context) error {
    userID, err := utils.GetUserIDFromContext(c)
    if err != nil {
        return err
    }

    out, err := h.workspaceUC.ListPublicWorkspaces(c.Request().Context(), userID)
    if err != nil {
        return handleUseCaseError(err)
    }
    return c.JSON(http.StatusOK, out)
}

// JoinPublicWorkspace implements ServerInterface.JoinPublicWorkspace
func (h *WorkspaceHandler) JoinPublicWorkspace(c echo.Context, id string) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}
	_, err = h.workspaceUC.JoinPublicWorkspace(c.Request().Context(), workspaceuc.JoinPublicWorkspaceInput{
		WorkspaceID: id,
		UserID:      userID,
	})
	if err != nil {
		return handleUseCaseError(err)
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "ワークスペースに参加しました"})
}

// AddMemberByEmail implements ServerInterface.AddMemberByEmail
func (h *WorkspaceHandler) AddMemberByEmail(c echo.Context, id string) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}
	var req openapi.AddMemberRequest
	if err := c.Bind(&req); err != nil {
		return utils.HandleBindError(err)
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var role string
	if req.Role != nil {
		role = string(*req.Role)
	}

	_, err = h.workspaceUC.AddMemberByEmail(c.Request().Context(), workspaceuc.AddMemberByEmailInput{
		WorkspaceID: id,
		Email:       string(req.Email),
		Role:        role,
		RequestedBy: userID,
	})
	if err != nil {
		return handleUseCaseError(err)
	}
	return c.JSON(http.StatusCreated, map[string]string{"message": "メンバーを追加しました"})
}

// DeleteWorkspace implements ServerInterface.DeleteWorkspace
func (h *WorkspaceHandler) DeleteWorkspace(c echo.Context, id string) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "コンテキストにユーザーIDが含まれていません")
	}

	input := workspaceuc.DeleteWorkspaceInput{
		ID:     id,
		UserID: userID,
	}

	_, err := h.workspaceUC.DeleteWorkspace(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// ListMembers implements ServerInterface.ListMembers
func (h *WorkspaceHandler) ListMembers(c echo.Context, id string) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "コンテキストにユーザーIDが含まれていません")
	}

	input := workspaceuc.ListMembersInput{
		WorkspaceID: id,
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
		return echo.NewHTTPError(http.StatusBadRequest, "ワークスペースIDは必須です")
	}

	var req AddMemberRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "リクエストボディが不正です")
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

// UpdateMemberRole implements ServerInterface.UpdateMemberRole
func (h *WorkspaceHandler) UpdateMemberRole(c echo.Context, id string, userId openapi_types.UUID) error {
	var req openapi.UpdateMemberRoleRequest
	if err := c.Bind(&req); err != nil {
		return utils.HandleBindError(err)
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := workspaceuc.UpdateMemberRoleInput{
		WorkspaceID: id,
		UserID:      userId.String(),
		Role:        string(req.Role),
	}

	member, err := h.workspaceUC.UpdateMemberRole(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, member)
}

// RemoveMember implements ServerInterface.RemoveMember
func (h *WorkspaceHandler) RemoveMember(c echo.Context, id string, userId openapi_types.UUID) error {
	removerID, ok := c.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	input := workspaceuc.RemoveMemberInput{
		WorkspaceID: id,
		UserID:      userId.String(),
		RemoverID:   removerID,
	}

	_, err := h.workspaceUC.RemoveMember(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
