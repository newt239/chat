package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	dmuc "github.com/newt239/chat/internal/usecase/dm"
)

type DMHandler struct {
	dmInteractor *dmuc.Interactor
}

func NewDMHandler(dmInteractor *dmuc.Interactor) *DMHandler {
	return &DMHandler{dmInteractor: dmInteractor}
}

type CreateDMRequest struct {
	UserID string `json:"userId" validate:"required"`
}

type CreateGroupDMRequest struct {
	UserIDs []string `json:"userIds" validate:"required,min=2,max=9"`
	Name    string   `json:"name"`
}

func (h *DMHandler) CreateDM(c echo.Context) error {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	var req CreateDMRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := dmuc.CreateDMInput{
		WorkspaceID:  workspaceID,
		UserID:       userID,
		TargetUserID: req.UserID,
	}

	dm, err := h.dmInteractor.CreateDM(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, dm)
}

func (h *DMHandler) CreateGroupDM(c echo.Context) error {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	var req CreateGroupDMRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := dmuc.CreateGroupDMInput{
		WorkspaceID: workspaceID,
		CreatorID:   userID,
		MemberIDs:   req.UserIDs,
		Name:        req.Name,
	}

	dm, err := h.dmInteractor.CreateGroupDM(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, dm)
}

func (h *DMHandler) ListDMs(c echo.Context) error {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	input := dmuc.ListDMsInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		RequestUserID: userID,
	}

	dms, err := h.dmInteractor.ListDMs(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, dms)
}
