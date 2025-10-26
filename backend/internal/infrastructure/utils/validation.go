package utils

import (
	"github.com/labstack/echo/v4"
)

// ValidateRequest はリクエストのバリデーションを統一して処理します
func ValidateRequest(c echo.Context, req interface{}) error {
	// リクエストのバインディング
	if err := c.Bind(req); err != nil {
		return HandleBindError(err)
	}

	// バリデーション
	if err := c.Validate(req); err != nil {
		return HandleValidationError(err)
	}

	return nil
}

// GetUserIDFromContext はコンテキストからユーザーIDを取得します
func GetUserIDFromContext(c echo.Context) (string, error) {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return "", HandleAuthError()
	}
	return userID, nil
}

// GetParamFromContext はパラメータを取得し、空の場合はエラーを返します
func GetParamFromContext(c echo.Context, paramName string) (string, error) {
	param := c.Param(paramName)
	if param == "" {
		return "", HandleParamError(paramName)
	}
	return param, nil
}
