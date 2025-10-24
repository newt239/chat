package handler

import (
	"errors"
	"net/http"

	domerr "github.com/example/chat/internal/domain/errors"
	"github.com/labstack/echo/v4"
)

// handleUseCaseError はユースケースのエラーをHTTPステータスコードに変換します
func handleUseCaseError(err error) error {
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
