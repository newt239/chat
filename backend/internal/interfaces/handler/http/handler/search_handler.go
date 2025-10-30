package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/infrastructure/utils"
	searchuc "github.com/newt239/chat/internal/usecase/search"
)

type SearchHandler struct {
	searchUC searchuc.SearchUseCase
}

func NewSearchHandler(searchUC searchuc.SearchUseCase) *SearchHandler {
	return &SearchHandler{searchUC: searchUC}
}

func (h *SearchHandler) SearchWorkspace(c echo.Context) error {
	workspaceID := c.Param("workspaceId")
	if workspaceID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ワークスペースIDは必須です")
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	query := c.QueryParam("q")
	filterParam := c.QueryParam("filter")
	pageParam := c.QueryParam("page")
	perPageParam := c.QueryParam("perPage")

	page := 1
	if pageParam != "" {
		if parsed, err := strconv.Atoi(pageParam); err == nil && parsed > 0 {
			page = parsed
		}
	}

	perPage := 0
	if perPageParam != "" {
		if parsed, err := strconv.Atoi(perPageParam); err == nil {
			perPage = parsed
		}
	}

	input := searchuc.WorkspaceSearchInput{
		WorkspaceID: workspaceID,
		RequesterID: userID,
		Query:       query,
		Filter:      searchuc.SearchFilter(filterParam).Normalize(),
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
