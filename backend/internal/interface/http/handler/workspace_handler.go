package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/example/chat/internal/usecase/workspace"
)

type WorkspaceHandler struct {
	workspaceUseCase workspace.WorkspaceUseCase
}

func NewWorkspaceHandler(workspaceUseCase workspace.WorkspaceUseCase) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceUseCase: workspaceUseCase,
	}
}

// GetWorkspaces godoc
// @Summary Get user workspaces
// @Description Get all workspaces that the authenticated user is a member of
// @Tags workspace
// @Produce json
// @Success 200 {object} workspace.GetWorkspacesOutput
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/workspaces [get]
// @Security BearerAuth
func (h *WorkspaceHandler) GetWorkspaces(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	output, err := h.workspaceUseCase.GetWorkspacesByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get workspaces"})
		return
	}

	c.JSON(http.StatusOK, output)
}

// GetWorkspace godoc
// @Summary Get workspace details
// @Description Get details of a specific workspace
// @Tags workspace
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 200 {object} workspace.GetWorkspaceOutput
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/workspaces/{id} [get]
// @Security BearerAuth
func (h *WorkspaceHandler) GetWorkspace(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Workspace ID is required"})
		return
	}

	input := workspace.GetWorkspaceInput{
		ID:     workspaceID,
		UserID: userID,
	}

	output, err := h.workspaceUseCase.GetWorkspace(c.Request.Context(), input)
	if err != nil {
		if err == workspace.ErrUnauthorized {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		if err == workspace.ErrWorkspaceNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get workspace"})
		return
	}

	c.JSON(http.StatusOK, output)
}

// CreateWorkspace godoc
// @Summary Create workspace
// @Description Create a new workspace
// @Tags workspace
// @Accept json
// @Produce json
// @Param request body CreateWorkspaceRequest true "Create workspace request"
// @Success 201 {object} workspace.CreateWorkspaceOutput
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/workspaces [post]
// @Security BearerAuth
func (h *WorkspaceHandler) CreateWorkspace(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	var req CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	input := workspace.CreateWorkspaceInput{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   userID,
	}

	output, err := h.workspaceUseCase.CreateWorkspace(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create workspace"})
		return
	}

	c.JSON(http.StatusCreated, output)
}

// UpdateWorkspace godoc
// @Summary Update workspace
// @Description Update workspace details (admin/owner only)
// @Tags workspace
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Param request body UpdateWorkspaceRequest true "Update workspace request"
// @Success 200 {object} workspace.UpdateWorkspaceOutput
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/workspaces/{id} [patch]
// @Security BearerAuth
func (h *WorkspaceHandler) UpdateWorkspace(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Workspace ID is required"})
		return
	}

	var req UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	input := workspace.UpdateWorkspaceInput{
		ID:          workspaceID,
		Name:        req.Name,
		Description: req.Description,
		IconURL:     req.IconURL,
		UserID:      userID,
	}

	output, err := h.workspaceUseCase.UpdateWorkspace(c.Request.Context(), input)
	if err != nil {
		if err == workspace.ErrUnauthorized {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		if err == workspace.ErrWorkspaceNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update workspace"})
		return
	}

	c.JSON(http.StatusOK, output)
}

// DeleteWorkspace godoc
// @Summary Delete workspace
// @Description Delete a workspace (owner only)
// @Tags workspace
// @Param id path string true "Workspace ID"
// @Success 200 {object} workspace.DeleteWorkspaceOutput
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/workspaces/{id} [delete]
// @Security BearerAuth
func (h *WorkspaceHandler) DeleteWorkspace(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Workspace ID is required"})
		return
	}

	input := workspace.DeleteWorkspaceInput{
		ID:     workspaceID,
		UserID: userID,
	}

	output, err := h.workspaceUseCase.DeleteWorkspace(c.Request.Context(), input)
	if err != nil {
		if err == workspace.ErrUnauthorized {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete workspace"})
		return
	}

	c.JSON(http.StatusOK, output)
}

// ListMembers godoc
// @Summary List workspace members
// @Description Get all members of a workspace
// @Tags workspace
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 200 {object} workspace.ListMembersOutput
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/workspaces/{id}/members [get]
// @Security BearerAuth
func (h *WorkspaceHandler) ListMembers(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Workspace ID is required"})
		return
	}

	input := workspace.ListMembersInput{
		WorkspaceID: workspaceID,
		RequesterID: userID,
	}

	output, err := h.workspaceUseCase.ListMembers(c.Request.Context(), input)
	if err != nil {
		if err == workspace.ErrUnauthorized {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to list members"})
		return
	}

	c.JSON(http.StatusOK, output)
}

// AddMember godoc
// @Summary Add workspace member
// @Description Add a user to a workspace (admin/owner only)
// @Tags workspace
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Param request body AddMemberRequest true "Add member request"
// @Success 200 {object} workspace.MemberActionOutput
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/workspaces/{id}/members [post]
// @Security BearerAuth
func (h *WorkspaceHandler) AddMember(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Workspace ID is required"})
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	input := workspace.AddMemberInput{
		WorkspaceID: workspaceID,
		UserID:      req.UserID,
		InviterID:   userID,
		Role:        req.Role,
	}

	output, err := h.workspaceUseCase.AddMember(c.Request.Context(), input)
	if err != nil {
		if err == workspace.ErrUnauthorized {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		if err == workspace.ErrInvalidRole {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to add member"})
		return
	}

	c.JSON(http.StatusOK, output)
}

// UpdateMemberRole godoc
// @Summary Update member role
// @Description Update a member's role in a workspace (owner only)
// @Tags workspace
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Param userId path string true "User ID"
// @Param request body UpdateMemberRoleRequest true "Update role request"
// @Success 200 {object} workspace.MemberActionOutput
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/workspaces/{id}/members/{userId} [patch]
// @Security BearerAuth
func (h *WorkspaceHandler) UpdateMemberRole(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	workspaceID := c.Param("id")
	targetUserID := c.Param("userId")
	if workspaceID == "" || targetUserID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Workspace ID and User ID are required"})
		return
	}

	var req UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	input := workspace.UpdateMemberRoleInput{
		WorkspaceID: workspaceID,
		UserID:      targetUserID,
		UpdaterID:   userID,
		Role:        req.Role,
	}

	output, err := h.workspaceUseCase.UpdateMemberRole(c.Request.Context(), input)
	if err != nil {
		if err == workspace.ErrUnauthorized {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		if err == workspace.ErrInvalidRole || err == workspace.ErrCannotChangeOwnerRole {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update member role"})
		return
	}

	c.JSON(http.StatusOK, output)
}

// RemoveMember godoc
// @Summary Remove workspace member
// @Description Remove a user from a workspace (admin/owner only)
// @Tags workspace
// @Param id path string true "Workspace ID"
// @Param userId path string true "User ID"
// @Success 200 {object} workspace.MemberActionOutput
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/workspaces/{id}/members/{userId} [delete]
// @Security BearerAuth
func (h *WorkspaceHandler) RemoveMember(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	workspaceID := c.Param("id")
	targetUserID := c.Param("userId")
	if workspaceID == "" || targetUserID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Workspace ID and User ID are required"})
		return
	}

	input := workspace.RemoveMemberInput{
		WorkspaceID: workspaceID,
		UserID:      targetUserID,
		RemoverID:   userID,
	}

	output, err := h.workspaceUseCase.RemoveMember(c.Request.Context(), input)
	if err != nil {
		if err == workspace.ErrUnauthorized {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		if err == workspace.ErrCannotRemoveOwner {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to remove member"})
		return
	}

	c.JSON(http.StatusOK, output)
}
