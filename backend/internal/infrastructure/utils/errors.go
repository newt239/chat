package utils

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	domerr "github.com/newt239/chat/internal/domain/errors"
)

// HandleUseCaseError はユースケースのエラーをHTTPステータスコードに変換します
func HandleUseCaseError(err error) error {
	if err == nil {
		return nil
	}

	// ドメインエラーの場合
	switch {
	case errors.Is(err, domerr.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "Resource not found")
	case errors.Is(err, domerr.ErrUnauthorized):
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	case errors.Is(err, domerr.ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	case errors.Is(err, domerr.ErrConflict):
		return echo.NewHTTPError(http.StatusConflict, "Conflict")
	case errors.Is(err, domerr.ErrValidation):
		return echo.NewHTTPError(http.StatusBadRequest, "Validation error")
	case errors.Is(err, domerr.ErrInternal):
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// その他のエラーは500として扱う
	return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
}

// HandleBindError はリクエストバインディングエラーを処理します
func HandleBindError(err error) error {
	if err == nil {
		return nil
	}
	return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
}

// HandleValidationError はバリデーションエラーを処理します
func HandleValidationError(err error) error {
	if err == nil {
		return nil
	}
	return echo.NewHTTPError(http.StatusBadRequest, err.Error())
}

// HandleAuthError は認証エラーを処理します
func HandleAuthError() error {
	return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
}

// HandleParamError はパラメータエラーを処理します
func HandleParamError(paramName string) error {
	return echo.NewHTTPError(http.StatusBadRequest, paramName+" is required")
}
