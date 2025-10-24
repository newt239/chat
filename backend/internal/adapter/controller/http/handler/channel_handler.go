package handler

import (
	"net/http"

	channeluc "github.com/example/chat/internal/usecase/channel"
	"github.com/labstack/echo/v4"
)

type ChannelHandler struct {
	channelUC channeluc.ChannelUseCase
}

func NewChannelHandler(channelUC channeluc.ChannelUseCase) *ChannelHandler {
	return &ChannelHandler{channelUC: channelUC}
}

// CreateChannelRequest はチャンネル作成リクエストの構造体です
type CreateChannelRequest struct {
	Name        string `json:"name" validate:"required,min=1"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
}

// UpdateChannelRequest はチャンネル更新リクエストの構造体です
type UpdateChannelRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1"`
	Description *string `json:"description,omitempty"`
	IsPrivate   *bool   `json:"is_private,omitempty"`
}

// ListChannels はチャンネル一覧を取得します
func (h *ChannelHandler) ListChannels(c echo.Context) error {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	input := channeluc.ListChannelsInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}

	channels, err := h.channelUC.ListChannels(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, channels)
}

// CreateChannel はチャンネルを作成します
func (h *ChannelHandler) CreateChannel(c echo.Context) error {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	var req CreateChannelRequest
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

	input := channeluc.CreateChannelInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Name:        req.Name,
		Description: description,
		IsPrivate:   req.IsPrivate,
	}

	channel, err := h.channelUC.CreateChannel(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusCreated, channel)
}
