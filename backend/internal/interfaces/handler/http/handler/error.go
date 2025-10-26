package handler

import (
	"github.com/newt239/chat/internal/infrastructure/utils"
)

// ErrorResponse represents a generic error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// handleUseCaseError はユースケースのエラーをHTTPステータスコードに変換します
// 共通のユーティリティを使用するように変更
func handleUseCaseError(err error) error {
	return utils.HandleUseCaseError(err)
}
