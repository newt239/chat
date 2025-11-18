package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/newt239/chat/internal/infrastructure/utils"
	"github.com/newt239/chat/internal/openapi_gen"
	searchuc "github.com/newt239/chat/internal/usecase/search"
)

type SearchHandler struct {
	searchUC searchuc.SearchUseCase
}

func NewSearchHandler(searchUC searchuc.SearchUseCase) *SearchHandler {
	return &SearchHandler{searchUC: searchUC}
}

// SearchWorkspace implements ServerInterface.SearchWorkspace
func (h *SearchHandler) SearchWorkspace(c echo.Context, workspaceId string, params openapi.SearchWorkspaceParams) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	filter := searchuc.SearchFilter("")
	if params.Filter != nil {
		filter = searchuc.SearchFilter(*params.Filter)
	}

	page := 1
	if params.Page != nil && *params.Page > 0 {
		page = *params.Page
	}

	perPage := 0
	if params.PerPage != nil {
		perPage = *params.PerPage
	}

	input := searchuc.WorkspaceSearchInput{
		WorkspaceID: workspaceId,
		RequesterID: userID,
		Query:       params.Q,
		Filter:      filter.Normalize(),
		Page:        page,
		PerPage:     perPage,
	}

	result, err := h.searchUC.SearchWorkspace(c.Request().Context(), input)
	if err != nil {
		switch err {
		case searchuc.ErrInvalidQuery:
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		case searchuc.ErrWorkspaceNotFound:
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		case searchuc.ErrUnauthorized:
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		default:
			return handleUseCaseError(err)
		}
	}

	return c.JSON(http.StatusOK, result)
}
