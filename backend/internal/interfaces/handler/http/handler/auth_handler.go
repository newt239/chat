package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/infrastructure/utils"
	openapi "github.com/newt239/chat/internal/openapi_gen"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

type AuthHandler struct {
	AuthUC authuc.AuthUseCase
}

// LogoutRequest はログアウトリクエストの構造体です
// 注意: OpenAPIスキーマに定義がないため、一時的に独自型を使用
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Register はユーザー登録を処理します
func (h *AuthHandler) Register(c echo.Context) error {
	var req openapi.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return utils.HandleBindError(err)
	}

	// Validation
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := authuc.RegisterInput{
		Email:       string(req.Email),
		Password:    req.Password,
		DisplayName: req.DisplayName,
	}

	output, err := h.AuthUC.Register(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusCreated, output)
}

// Login はユーザー認証を処理します
func (h *AuthHandler) Login(c echo.Context) error {
	var req openapi.LoginRequest
	if err := c.Bind(&req); err != nil {
		return utils.HandleBindError(err)
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := authuc.LoginInput{
		Email:    string(req.Email),
		Password: req.Password,
	}

	output, err := h.AuthUC.Login(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, output)
}

// Refresh はトークンのリフレッシュを処理します
func (h *AuthHandler) Refresh(c echo.Context) error {
	var req openapi.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return utils.HandleBindError(err)
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := authuc.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
	}

	output, err := h.AuthUC.RefreshToken(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, output)
}

// Logout はログアウトを処理します
func (h *AuthHandler) Logout(c echo.Context) error {
	var req LogoutRequest
	if err := c.Bind(&req); err != nil {
		return utils.HandleBindError(err)
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

	output, err := h.AuthUC.Logout(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, output)
}
