package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/newt239/chat/internal/infrastructure/utils"
	openapi "github.com/newt239/chat/internal/openapi_gen"
	useruc "github.com/newt239/chat/internal/usecase/user"
)

type UserHandler struct {
	UC useruc.UseCase
}

func (h *UserHandler) GetMe(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return utils.HandleAuthError()
	}

	out, err := h.UC.GetMe(c.Request().Context(), userID)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, out)
}

func (h *UserHandler) UpdateMe(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return utils.HandleAuthError()
	}

	var req openapi.UpdateMeRequest
	if err := c.Bind(&req); err != nil {
		return utils.HandleBindError(err)
	}

	if err := c.Validate(&req); err != nil {
		return utils.HandleValidationError(err)
	}

	out, err := h.UC.UpdateMe(c.Request().Context(), useruc.UpdateMeInput{
		UserID:      userID,
		DisplayName: req.DisplayName,
		Bio:         req.Bio,
		AvatarURL:   req.AvatarUrl,
	})
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, out)
}

func (h *UserHandler) UpdatePassword(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return utils.HandleAuthError()
	}

	var req openapi.UpdatePasswordRequest
	if err := c.Bind(&req); err != nil {
		return utils.HandleBindError(err)
	}

	if err := c.Validate(&req); err != nil {
		return utils.HandleValidationError(err)
	}

	err := h.UC.UpdatePassword(c.Request().Context(), useruc.UpdatePasswordInput{
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	})
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, map[string]bool{"success": true})
}

func (h *UserHandler) DeleteMe(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return utils.HandleAuthError()
	}

	if err := h.UC.DeleteMe(c.Request().Context(), userID); err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, map[string]bool{"success": true})
}
