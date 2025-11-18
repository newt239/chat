package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/newt239/chat/internal/infrastructure/utils"
	openapi "github.com/newt239/chat/internal/openapi_gen"
	dmuc "github.com/newt239/chat/internal/usecase/dm"
)

type DMHandler struct {
	DMInteractor *dmuc.Interactor
}

// CreateDM implements ServerInterface.CreateDM
func (h *DMHandler) CreateDM(c echo.Context, id openapi_types.UUID) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	var req openapi.CreateDMRequest
	if err := c.Bind(&req); err != nil {
		return utils.HandleBindError(err)
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := dmuc.CreateDMInput{
		WorkspaceID:  id.String(),
		UserID:       userID,
		TargetUserID: req.UserId.String(),
	}

	dm, err := h.DMInteractor.CreateDM(c.Request().Context(), input)
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

	var req openapi.CreateGroupDMRequest
	if err := c.Bind(&req); err != nil {
		return utils.HandleBindError(err)
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userIDs := make([]string, len(req.UserIds))
	for i, id := range req.UserIds {
		userIDs[i] = id.String()
	}

	var name string
	if req.Name != nil {
		name = *req.Name
	}

	input := dmuc.CreateGroupDMInput{
		WorkspaceID: id.String(),
		CreatorID:   userID,
		MemberIDs:   userIDs,
		Name:        name,
	}

	dm, err := h.DMInteractor.CreateGroupDM(c.Request().Context(), input)
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

	dms, err := h.DMInteractor.ListDMs(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, dms)
}
