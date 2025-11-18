package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/infrastructure/utils"
	"github.com/newt239/chat/internal/openapi_gen"
	linkuc "github.com/newt239/chat/internal/usecase/link"
)

type LinkHandler struct {
	linkUC linkuc.LinkUseCase
}

func NewLinkHandler(linkUC linkuc.LinkUseCase) *LinkHandler {
	return &LinkHandler{linkUC: linkUC}
}

// FetchOGP はOGP情報を取得します
func (h *LinkHandler) FetchOGP(c echo.Context) error {
	var req openapi.FetchOGPRequest
	if err := c.Bind(&req); err != nil {
		return utils.HandleBindError(err)
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := linkuc.FetchOGPInput{
		URL: req.Url,
	}

	ogp, err := h.linkUC.FetchOGP(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, ogp)
}
