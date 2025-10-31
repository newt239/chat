package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/infrastructure/utils"
	channeluc "github.com/newt239/chat/internal/usecase/channel"
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
		return echo.NewHTTPError(http.StatusBadRequest, "ワークスペースIDは必須です")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
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
		return echo.NewHTTPError(http.StatusBadRequest, "ワークスペースIDは必須です")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	var req CreateChannelRequest
	if err := c.Bind(&req); err != nil {
		return utils.HandleBindError(err)
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

// UpdateChannel はチャンネル情報を更新します
func (h *ChannelHandler) UpdateChannel(c echo.Context) error {
    channelID := c.Param("channelId")
    if channelID == "" {
        return echo.NewHTTPError(http.StatusBadRequest, "チャンネルIDは必須です")
    }

    userID, ok := c.Get("userID").(string)
    if !ok {
        return utils.HandleAuthError()
    }

    var req UpdateChannelRequest
    if err := c.Bind(&req); err != nil {
        return utils.HandleBindError(err)
    }

    input := channeluc.UpdateChannelInput{
        ChannelID:   channelID,
        UserID:      userID,
        Name:        req.Name,
        Description: req.Description,
        IsPrivate:   req.IsPrivate,
    }

    ch, err := h.channelUC.UpdateChannel(c.Request().Context(), input)
    if err != nil {
        return handleUseCaseError(err)
    }
    return c.JSON(http.StatusOK, ch)
}
