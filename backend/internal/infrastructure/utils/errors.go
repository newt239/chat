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
		return echo.NewHTTPError(http.StatusNotFound, "指定されたリソースが見つかりません")
	case errors.Is(err, domerr.ErrUnauthorized):
		return echo.NewHTTPError(http.StatusUnauthorized, "操作を実行する権限がありません")
	case errors.Is(err, domerr.ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, "アクセスが禁止されています")
	case errors.Is(err, domerr.ErrConflict):
		return echo.NewHTTPError(http.StatusConflict, "処理が競合しました")
	case errors.Is(err, domerr.ErrValidation):
		return echo.NewHTTPError(http.StatusBadRequest, "入力値が条件を満たしていません")
	case errors.Is(err, domerr.ErrInternal):
		return echo.NewHTTPError(http.StatusInternalServerError, "サーバー内部でエラーが発生しました")
	}

	// その他のエラーは500として扱う
	return echo.NewHTTPError(http.StatusInternalServerError, "サーバー内部でエラーが発生しました")
}

// HandleBindError はリクエストバインディングエラーを処理します
func HandleBindError(err error) error {
	if err == nil {
		return nil
	}
	return echo.NewHTTPError(http.StatusBadRequest, "リクエストボディが不正です: "+err.Error())
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
	return echo.NewHTTPError(http.StatusUnauthorized, "コンテキストにユーザーIDが含まれていません")
}

// HandleParamError はパラメータエラーを処理します
func HandleParamError(paramName string) error {
	return echo.NewHTTPError(http.StatusBadRequest, paramName+"は必須です")
}
