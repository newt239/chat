package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

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

// CreateDM implements ServerInterface.CreateDM
func (h *DMHandler) CreateDM(c echo.Context, id openapi_types.UUID) error {
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
		WorkspaceID:  id.String(),
		UserID:       userID,
		TargetUserID: req.UserID,
	}

	dm, err := h.dmInteractor.CreateDM(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, dm)
}

// CreateGroupDM implements ServerInterface.CreateGroupDM
func (h *DMHandler) CreateGroupDM(c echo.Context, id openapi_types.UUID) error {
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
		WorkspaceID: id.String(),
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

// ListDMs implements ServerInterface.ListDMs
func (h *DMHandler) ListDMs(c echo.Context, id openapi_types.UUID) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	input := dmuc.ListDMsInput{
		WorkspaceID: id.String(),
		UserID:      userID,
	}

	dms, err := h.dmInteractor.ListDMs(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, dms)
}
