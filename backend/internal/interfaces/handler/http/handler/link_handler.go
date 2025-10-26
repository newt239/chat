package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	linkuc "github.com/newt239/chat/internal/usecase/link"
)

type LinkHandler struct {
	linkUC linkuc.LinkUseCase
}

func NewLinkHandler(linkUC linkuc.LinkUseCase) *LinkHandler {
	return &LinkHandler{linkUC: linkUC}
}

// FetchOGPRequest はOGP取得リクエストの構造体です
type FetchOGPRequest struct {
	URL string `json:"url" validate:"required,url"`
}

// FetchOGP はOGP情報を取得します
func (h *LinkHandler) FetchOGP(c echo.Context) error {
	var req FetchOGPRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := linkuc.FetchOGPInput{
		URL: req.URL,
	}

	ogp, err := h.linkUC.FetchOGP(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, ogp)
}
