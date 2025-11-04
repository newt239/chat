package utils

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	domerr "github.com/newt239/chat/internal/domain/errors"
	"github.com/newt239/chat/internal/infrastructure/logger"
	"go.uber.org/zap"
)

// HandleUseCaseError はユースケースのエラーをHTTPステータスコードに変換します
func HandleUseCaseError(err error) error {
	if err == nil {
		return nil
	}

	log := logger.Get()

	// ドメインエラーの場合
	switch {
	case errors.Is(err, domerr.ErrNotFound):
		log.Debug("リソースが見つかりません", zap.Error(err))
		return echo.NewHTTPError(http.StatusNotFound, "指定されたリソースが見つかりません")
	case errors.Is(err, domerr.ErrUnauthorized):
		log.Warn("認証エラー", zap.Error(err))
		return echo.NewHTTPError(http.StatusUnauthorized, "操作を実行する権限がありません")
	case errors.Is(err, domerr.ErrForbidden):
		log.Warn("アクセス禁止エラー", zap.Error(err))
		return echo.NewHTTPError(http.StatusForbidden, "アクセスが禁止されています")
	case errors.Is(err, domerr.ErrConflict):
		log.Warn("競合エラー", zap.Error(err))
		return echo.NewHTTPError(http.StatusConflict, "処理が競合しました")
	case errors.Is(err, domerr.ErrValidation):
		log.Debug("バリデーションエラー", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "入力値が条件を満たしていません")
	case errors.Is(err, domerr.ErrInternal):
		log.Error("内部エラー", zap.Error(err), zap.Stack("stacktrace"))
		return echo.NewHTTPError(http.StatusInternalServerError, "サーバー内部でエラーが発生しました")
	}

	// その他のエラーは500として扱う
	log.Error("予期しないエラー", zap.Error(err), zap.Stack("stacktrace"))
	return echo.NewHTTPError(http.StatusInternalServerError, "サーバー内部でエラーが発生しました")
}

// HandleBindError はリクエストバインディングエラーを処理します
func HandleBindError(err error) error {
	if err == nil {
		return nil
	}
	log := logger.Get()
	log.Debug("リクエストバインディングエラー", zap.Error(err))
	return echo.NewHTTPError(http.StatusBadRequest, "リクエストボディが不正です: "+err.Error())
}

// HandleValidationError はバリデーションエラーを処理します
func HandleValidationError(err error) error {
	if err == nil {
		return nil
	}
	log := logger.Get()
	log.Debug("バリデーションエラー", zap.Error(err))
	return echo.NewHTTPError(http.StatusBadRequest, err.Error())
}

// HandleAuthError は認証エラーを処理します
func HandleAuthError() error {
	log := logger.Get()
	log.Warn("認証エラー: コンテキストにユーザーIDが含まれていません")
	return echo.NewHTTPError(http.StatusUnauthorized, "コンテキストにユーザーIDが含まれていません")
}

// HandleParamError はパラメータエラーを処理します
func HandleParamError(paramName string) error {
	log := logger.Get()
	log.Debug("パラメータエラー", zap.String("param", paramName))
	return echo.NewHTTPError(http.StatusBadRequest, paramName+"は必須です")
}
