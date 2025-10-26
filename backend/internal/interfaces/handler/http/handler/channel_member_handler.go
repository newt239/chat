package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/newt239/chat/internal/usecase/channelmember"
)

type ChannelMemberHandler struct {
	channelMemberUseCase channelmember.ChannelMemberUseCase
}

func NewChannelMemberHandler(channelMemberUseCase channelmember.ChannelMemberUseCase) *ChannelMemberHandler {
	return &ChannelMemberHandler{channelMemberUseCase: channelMemberUseCase}
}

type InviteMemberRequest struct {
	UserID string  `json:"userId"`
	Role   *string `json:"role"`
}

func (r *InviteMemberRequest) Validate() error {
	if strings.TrimSpace(r.UserID) == "" {
		return errors.New("userId is required")
	}
	if r.Role != nil && (*r.Role != "member" && *r.Role != "admin") {
		return errors.New("role must be either 'member' or 'admin'")
	}
	return nil
}

type ChannelMemberUpdateRoleRequest struct {
	Role string `json:"role"`
}

func (r *ChannelMemberUpdateRoleRequest) Validate() error {
	if strings.TrimSpace(r.Role) == "" {
		return errors.New("role is required")
	}
	return nil
}

type SuccessResponse struct {
	Success bool `json:"success"`
}

// ListMembers godoc
// @Summary List channel members
// @Description Returns members in the specified channel
// @Tags channel-member
// @Produce json
// @Param channelId path string true "Channel ID"
// @Success 200 {object} channelmember.MemberListOutput
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/channels/{channelId}/members [get]
// @Security BearerAuth
func (h *ChannelMemberHandler) ListMembers(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	output, err := h.channelMemberUseCase.ListMembers(c.Request().Context(), channelmember.ListMembersInput{
		ChannelID: channelID,
		UserID:    userID,
	})
	if err != nil {
		switch err {
		case channelmember.ErrChannelNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to list members"})
		}
	}

	return c.JSON(http.StatusOK, output)
}

// InviteMember godoc
// @Summary Invite a member to a channel
// @Description Invites a user to a channel with a specified role
// @Tags channel-member
// @Accept json
// @Produce json
// @Param channelId path string true "Channel ID"
// @Param request body InviteMemberRequest true "Invite member request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/channels/{channelId}/members [post]
// @Security BearerAuth
func (h *ChannelMemberHandler) InviteMember(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	var req InviteMemberRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	role := "member"
	if req.Role != nil {
		role = *req.Role
	}

	err := h.channelMemberUseCase.InviteMember(c.Request().Context(), channelmember.InviteMemberInput{
		ChannelID:    channelID,
		OperatorID:   userID,
		TargetUserID: req.UserID,
		Role:         role,
	})
	if err != nil {
		switch err {
		case channelmember.ErrChannelNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrUserNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case channelmember.ErrAlreadyMember:
			return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		case channelmember.ErrInvalidRole:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to invite member"})
		}
	}

	return c.JSON(http.StatusOK, SuccessResponse{Success: true})
}

// JoinPublicChannel godoc
// @Summary Join a public channel
// @Description Allows a user to self-join a public channel
// @Tags channel-member
// @Produce json
// @Param channelId path string true "Channel ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/channels/{channelId}/members/self [post]
// @Security BearerAuth
func (h *ChannelMemberHandler) JoinPublicChannel(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	err := h.channelMemberUseCase.JoinPublicChannel(c.Request().Context(), channelmember.JoinChannelInput{
		ChannelID: channelID,
		UserID:    userID,
	})
	if err != nil {
		switch err {
		case channelmember.ErrChannelNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrChannelNotPublic:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case channelmember.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to join channel"})
		}
	}

	return c.JSON(http.StatusOK, SuccessResponse{Success: true})
}

// UpdateMemberRole godoc
// @Summary Update member role
// @Description Updates the role of a channel member
// @Tags channel-member
// @Accept json
// @Produce json
// @Param channelId path string true "Channel ID"
// @Param userId path string true "User ID"
// @Param request body ChannelMemberUpdateRoleRequest true "Update member role request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/channels/{channelId}/members/{userId}/role [patch]
// @Security BearerAuth
func (h *ChannelMemberHandler) UpdateMemberRole(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	targetUserID := c.Param("userId")
	if targetUserID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "User ID is required"})
	}

	var req ChannelMemberUpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	err := h.channelMemberUseCase.UpdateMemberRole(c.Request().Context(), channelmember.UpdateMemberRoleInput{
		ChannelID:    channelID,
		OperatorID:   userID,
		TargetUserID: targetUserID,
		Role:         req.Role,
	})
	if err != nil {
		switch err {
		case channelmember.ErrChannelNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrNotMember:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case channelmember.ErrInvalidRole:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		case channelmember.ErrLastAdminRemoval:
			return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update member role"})
		}
	}

	return c.JSON(http.StatusOK, SuccessResponse{Success: true})
}

// RemoveMember godoc
// @Summary Remove a member from a channel
// @Description Removes a user from a channel
// @Tags channel-member
// @Produce json
// @Param channelId path string true "Channel ID"
// @Param userId path string true "User ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/channels/{channelId}/members/{userId} [delete]
// @Security BearerAuth
func (h *ChannelMemberHandler) RemoveMember(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	targetUserID := c.Param("userId")
	if targetUserID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "User ID is required"})
	}

	err := h.channelMemberUseCase.RemoveMember(c.Request().Context(), channelmember.RemoveMemberInput{
		ChannelID:    channelID,
		OperatorID:   userID,
		TargetUserID: targetUserID,
	})
	if err != nil {
		switch err {
		case channelmember.ErrChannelNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrNotMember:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case channelmember.ErrLastAdminRemoval:
			return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to remove member"})
		}
	}

	return c.JSON(http.StatusOK, SuccessResponse{Success: true})
}

// LeaveChannel godoc
// @Summary Leave a channel
// @Description Allows a user to leave a channel
// @Tags channel-member
// @Produce json
// @Param channelId path string true "Channel ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/channels/{channelId}/members/self [delete]
// @Security BearerAuth
func (h *ChannelMemberHandler) LeaveChannel(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	err := h.channelMemberUseCase.LeaveChannel(c.Request().Context(), channelmember.LeaveChannelInput{
		ChannelID: channelID,
		UserID:    userID,
	})
	if err != nil {
		switch err {
		case channelmember.ErrChannelNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrNotMember:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrLastAdminRemoval:
			return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to leave channel"})
		}
	}

	return c.JSON(http.StatusOK, SuccessResponse{Success: true})
}
