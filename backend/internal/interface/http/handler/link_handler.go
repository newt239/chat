package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/example/chat/internal/usecase/link"
)

type LinkHandler struct {
	linkUseCase link.LinkUseCase
}

func NewLinkHandler(linkUseCase link.LinkUseCase) *LinkHandler {
	return &LinkHandler{
		linkUseCase: linkUseCase,
	}
}

// FetchOGP OGP情報取得
func (h *LinkHandler) FetchOGP(c echo.Context) error {
	var input link.FetchOGPInput
	if err := c.Bind(&input); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// URLの検証
	if input.URL == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "URL is required"})
		return echo.NewHTTPError(http.StatusBadRequest, "URL is required")
	}

	output, err := h.linkUseCase.FetchOGP(c.Request().Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return err
	}

	return c.JSON(http.StatusOK, output)
}
