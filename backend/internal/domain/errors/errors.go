package errors

import "errors"

var (
	ErrInvalidCredentials = errors.New("メールアドレスまたはパスワードが正しくありません")
	ErrUserAlreadyExists  = errors.New("ユーザーはすでに登録されています")
	ErrInvalidToken       = errors.New("トークンが無効または期限切れです")
	ErrSessionNotFound    = errors.New("セッションが見つかりません")
	ErrNotFound           = errors.New("指定されたリソースが見つかりません")
	ErrMessageNotFound    = errors.New("メッセージが見つかりません")
	ErrChannelNotFound    = errors.New("チャンネルが見つかりません")
	ErrUnauthorized       = errors.New("操作を実行する権限がありません")
	ErrForbidden          = errors.New("アクセスが禁止されています")
	ErrInvalidInput       = errors.New("入力内容が不正です")
	ErrConflict           = errors.New("処理が競合しました")
	ErrValidation         = errors.New("入力値が条件を満たしていません")
	ErrInternal           = errors.New("サーバー内部でエラーが発生しました")
)
