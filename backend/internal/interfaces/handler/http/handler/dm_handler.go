package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/infrastructure/utils"
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
		return echo.NewHTTPError(http.StatusBadRequest, "ワークスペースIDは必須です")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	var req CreateDMRequest
	if err := c.Bind(&req); err != nil {
		return utils.HandleBindError(err)
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
		return echo.NewHTTPError(http.StatusBadRequest, "ワークスペースIDは必須です")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	var req CreateGroupDMRequest
	if err := c.Bind(&req); err != nil {
		return utils.HandleBindError(err)
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
		return echo.NewHTTPError(http.StatusBadRequest, "ワークスペースIDは必須です")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	input := dmuc.ListDMsInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}

	dms, err := h.dmInteractor.ListDMs(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, dms)
}
