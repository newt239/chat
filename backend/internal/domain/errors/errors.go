package errors

import (
	"errors"
	"fmt"
)

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
	ErrUnauthorized         = errors.New("操作を実行する権限がありません")
	ErrForbidden            = errors.New("アクセスが禁止されています")
	ErrInvalidInput         = errors.New("入力内容が不正です")
	ErrConflict             = errors.New("処理が競合しました")
	ErrValidation           = errors.New("入力値が条件を満たしていません")
	ErrInternal             = errors.New("サーバー内部でエラーが発生しました")
)

// 見つからない
var (
	ErrNotFound              = errors.New("指定されたリソースが見つかりません")
	ErrWorkspaceNotFound     = errors.New("ワークスペースが見つかりません")
	ErrUserNotFound          = errors.New("ユーザーが見つかりません")
	ErrChannelNotFound       = errors.New("チャンネルが見つかりません")
	ErrMessageNotFound       = errors.New("メッセージが見つかりません")
	ErrParentMessageNotFound = errors.New("返信先のメッセージが見つかりません")
	ErrAttachmentNotFound    = errors.New("添付ファイルが見つかりません")
)

// 状態が操作を許さない
var (
	ErrChannelArchived  = errors.New("アーカイブされたチャンネルには投稿できません")
	ErrNotChannelMember = errors.New("チャンネルに参加すると投稿できます")
)

// 一意制約に当たった
var (
	ErrAlreadyMember     = errors.New("既にメンバーです")
	ErrPinExists         = errors.New("このメッセージは既にピン留めされています")
	ErrReactionExists    = errors.New("同じリアクションが既に追加されています")
	ErrBookmarkExists    = errors.New("このメッセージは既にブックマークされています")
	ErrWorkspaceIDExists = errors.New("このワークスペースIDは既に使用されています")
)

// 入力が条件を満たさない
var (
	ErrInvalidRole     = fmt.Errorf("%w: 指定できないロールです", ErrValidation)
	ErrInvalidTimeZone = fmt.Errorf("%w: タイムゾーンの指定が正しくありません", ErrValidation)
)
