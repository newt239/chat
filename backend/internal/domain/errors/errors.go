package errors

import "errors"

var (
	ErrInvalidCredentials = errors.New("メールアドレスまたはパスワードが正しくありません")
	ErrUserAlreadyExists  = errors.New("ユーザーはすでに登録されています")
	ErrInvalidToken       = errors.New("トークンが無効または期限切れです")
	ErrSessionNotFound    = errors.New("セッションが見つかりません")
	// 未登録のメールアドレスは招待がなければアカウントを作れない
	ErrInvitationRequired   = errors.New("このメールアドレスは招待されていません")
	ErrInvitationNotFound   = errors.New("招待が見つからないか、有効期限が切れています")
	ErrEmailNotVerified     = errors.New("メールアドレスが確認されていない Google アカウントです")
	ErrPasswordAuthDisabled = errors.New("パスワードによるログインは無効です")
	ErrGoogleAuthDisabled   = errors.New("このサーバーでは Google ログインが設定されていません")
	ErrSignupDisabled       = errors.New("このワークスペースでは新規登録を受け付けていません")
	ErrNotFound             = errors.New("指定されたリソースが見つかりません")
	ErrMessageNotFound      = errors.New("メッセージが見つかりません")
	ErrChannelNotFound      = errors.New("チャンネルが見つかりません")
	ErrChannelArchived      = errors.New("アーカイブされたチャンネルには投稿できません")
	ErrUnauthorized         = errors.New("操作を実行する権限がありません")
	ErrForbidden            = errors.New("アクセスが禁止されています")
	ErrInvalidInput         = errors.New("入力内容が不正です")
	ErrConflict             = errors.New("処理が競合しました")
	ErrValidation           = errors.New("入力値が条件を満たしていません")
	ErrInternal             = errors.New("サーバー内部でエラーが発生しました")
)
