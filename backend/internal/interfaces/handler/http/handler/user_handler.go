package handler

import (
    "net/http"

    "github.com/labstack/echo/v4"
    "github.com/newt239/chat/internal/infrastructure/utils"
    useruc "github.com/newt239/chat/internal/usecase/user"
)

type UserHandler struct {
    uc useruc.UseCase
}

func NewUserHandler(uc useruc.UseCase) *UserHandler {
    return &UserHandler{uc: uc}
}

type UpdateMeRequest struct {
    DisplayName *string `json:"display_name"`
    Bio         *string `json:"bio"`
    AvatarURL   *string `json:"avatar_url"`
}

func (h *UserHandler) UpdateMe(c echo.Context) error {
    userID, ok := c.Get("userID").(string)
    if !ok || userID == "" {
        return utils.HandleAuthError()
    }

    var req UpdateMeRequest
    if err := c.Bind(&req); err != nil {
        return utils.HandleBindError(err)
    }

    out, err := h.uc.UpdateMe(c.Request().Context(), useruc.UpdateMeInput{
        UserID:      userID,
        DisplayName: req.DisplayName,
        Bio:         req.Bio,
        AvatarURL:   req.AvatarURL,
    })
    if err != nil {
        return handleUseCaseError(err)
    }

    return c.JSON(http.StatusOK, out)
}


