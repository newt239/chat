package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/example/chat/internal/usecase/channel"
)

type ChannelHandler struct {
	channelUseCase channel.ChannelUseCase
}

func NewChannelHandler(channelUseCase channel.ChannelUseCase) *ChannelHandler {
	return &ChannelHandler{
		channelUseCase: channelUseCase,
	}
}

// ListChannels godoc
// @Summary List channels in a workspace
// @Description Returns channels accessible to the authenticated user within the workspace
// @Tags channel
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 200 {array} channel.ChannelOutput
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/workspaces/{id}/channels [get]
// @Security BearerAuth
func (h *ChannelHandler) ListChannels(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Workspace ID is required"})
		return err
	}

	channels, err := h.channelUseCase.ListChannels(c.Request().Context(), channel.ListChannelsInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		switch err {
		case channel.ErrWorkspaceNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channel.ErrUnauthorized:
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to list channels"})
		}
		return err
	}

	return c.JSON(http.StatusOK, channels)
}

// CreateChannel godoc
// @Summary Create a channel
// @Description Creates a new channel in the specified workspace
// @Tags channel
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Param request body CreateChannelRequest true "Create channel request"
// @Success 201 {object} channel.ChannelOutput
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/workspaces/{id}/channels [post]
// @Security BearerAuth
func (h *ChannelHandler) CreateChannel(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Workspace ID is required"})
		return err
	}

	var req CreateChannelRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	channelOutput, err := h.channelUseCase.CreateChannel(c.Request().Context(), channel.CreateChannelInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		IsPrivate:   req.IsPrivate,
	})
	if err != nil {
		switch err {
		case channel.ErrWorkspaceNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channel.ErrUnauthorized:
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create channel"})
		}
		return err
	}

	return c.JSON(http.StatusCreated, channelOutput)
}
