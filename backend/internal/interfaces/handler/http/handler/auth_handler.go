package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	authuc "github.com/newt239/chat/internal/usecase/auth"
    "github.com/newt239/chat/internal/infrastructure/utils"
)

type AuthHandler struct {
	authUC authuc.AuthUseCase
}

func NewAuthHandler(authUC authuc.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

// RegisterRequest はユーザー登録リクエストの構造体です
type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	DisplayName string `json:"display_name" validate:"required,min=1"`
}

// LoginRequest はログインリクエストの構造体です
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenRequest はトークンリフレッシュリクエストの構造体です
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest はログアウトリクエストの構造体です
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Register はユーザー登録を処理します
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validation
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := authuc.RegisterInput{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	}

	output, err := h.authUC.Register(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusCreated, output)
}

// Login はユーザー認証を処理します
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := authuc.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	output, err := h.authUC.Login(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, output)
}

// RefreshToken はトークンのリフレッシュを処理します
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := authuc.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
	}

	output, err := h.authUC.RefreshToken(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, output)
}

// Logout はログアウトを処理します
func (h *AuthHandler) Logout(c echo.Context) error {
	var req LogoutRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

    // コンテキストからユーザーIDを取得（統一ヘルパー）
    userID, err := utils.GetUserIDFromContext(c)
    if err != nil {
        return err
    }

	input := authuc.LogoutInput{
		UserID:       userID,
		RefreshToken: req.RefreshToken,
	}

	output, err := h.authUC.Logout(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, output)
}
