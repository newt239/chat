package handler

import (
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/infrastructure/utils"
	"github.com/newt239/chat/internal/openapi_gen"
	channeluc "github.com/newt239/chat/internal/usecase/channel"
)

type ChannelHandler struct {
	ChannelUC channeluc.ChannelUseCase
}

// ListChannels implements ServerInterface.ListChannels
func (h *ChannelHandler) ListChannels(c echo.Context, id string) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	input := channeluc.ListChannelsInput{
		WorkspaceID: id,
		UserID:      userID,
	}

	channels, err := h.ChannelUC.ListChannels(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, channels)
}

// CreateChannel implements ServerInterface.CreateChannel
func (h *ChannelHandler) CreateChannel(c echo.Context, id string) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	var req openapi.CreateChannelRequest
	if err := c.Bind(&req); err != nil {
		return utils.HandleBindError(err)
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var isPrivate bool
	if req.IsPrivate != nil {
		isPrivate = *req.IsPrivate
	}

	input := channeluc.CreateChannelInput{
		WorkspaceID: id,
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		IsPrivate:   isPrivate,
	}

	channel, err := h.ChannelUC.CreateChannel(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusCreated, channel)
}

// UpdateChannel はチャンネル情報を更新します (ServerInterface用)
func (h *ChannelHandler) UpdateChannel(c echo.Context, channelId openapi_types.UUID) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	var req openapi.UpdateChannelRequest
	if err := c.Bind(&req); err != nil {
		return utils.HandleBindError(err)
	}

	input := channeluc.UpdateChannelInput{
		ChannelID:   channelId.String(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		IsPrivate:   req.IsPrivate,
	}

	ch, err := h.ChannelUC.UpdateChannel(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}
	return c.JSON(http.StatusOK, ch)
}
