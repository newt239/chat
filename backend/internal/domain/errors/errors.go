package errors

import "errors"

// 種類を表す番兵。個々のエラーは New で種類を持たせ、rpc はこの種類だけで応答コードを決める
var (
	ErrNotFound           = errors.New("指定されたリソースが見つかりません")
	ErrUnauthenticated    = errors.New("認証されていません")
	ErrUnauthorized       = errors.New("操作を実行する権限がありません")
	ErrAlreadyExists      = errors.New("既に存在します")
	ErrValidation         = errors.New("入力値が条件を満たしていません")
	ErrFailedPrecondition = errors.New("現在の状態ではこの操作を行えません")
)

type kindError struct {
	kind error
	msg  string
}

func (e *kindError) Error() string { return e.msg }

func (e *kindError) Unwrap() error { return e.kind }

// New は errors.Is で kind と一致し、メッセージは msg だけのエラーを作ります
func New(kind error, msg string) error {
	return &kindError{kind: kind, msg: msg}
}

var (
	ErrInvalidCredentials = New(ErrUnauthenticated, "メールアドレスまたはパスワードが正しくありません")
	ErrInvalidToken       = New(ErrUnauthenticated, "トークンが無効または期限切れです")

	ErrForbidden          = New(ErrUnauthorized, "アクセスが禁止されています")
	ErrNotChannelMember   = New(ErrUnauthorized, "チャンネルに参加すると投稿できます")
	ErrInvitationRequired = New(ErrUnauthorized, "このメールアドレスは招待されていません")
	ErrEmailNotVerified   = New(ErrUnauthorized, "メールアドレスが確認されていない Google アカウントです")

	ErrWorkspaceNotFound     = New(ErrNotFound, "ワークスペースが見つかりません")
	ErrUserNotFound          = New(ErrNotFound, "ユーザーが見つかりません")
	ErrChannelNotFound       = New(ErrNotFound, "チャンネルが見つかりません")
	ErrMessageNotFound       = New(ErrNotFound, "メッセージが見つかりません")
	ErrParentMessageNotFound = New(ErrNotFound, "返信先のメッセージが見つかりません")
	ErrAttachmentNotFound    = New(ErrNotFound, "添付ファイルが見つかりません")
	ErrInvitationNotFound    = New(ErrNotFound, "招待が見つからないか、有効期限が切れています")

	ErrUserAlreadyExists     = New(ErrAlreadyExists, "ユーザーはすでに登録されています")
	ErrAlreadyMember         = New(ErrAlreadyExists, "既にメンバーです")
	ErrPinExists             = New(ErrAlreadyExists, "このメッセージは既にピン留めされています")
	ErrReactionExists        = New(ErrAlreadyExists, "同じリアクションが既に追加されています")
	ErrBookmarkExists        = New(ErrAlreadyExists, "このメッセージは既にブックマークされています")
	ErrWorkspaceIDExists     = New(ErrAlreadyExists, "このワークスペースIDは既に使用されています")
	ErrCustomEmojiNameExists = New(ErrAlreadyExists, "同じ名前のカスタム絵文字がすでにあります")

	ErrPasswordAuthDisabled = New(ErrFailedPrecondition, "パスワードによるログインは無効です")
	ErrGoogleAuthDisabled   = New(ErrFailedPrecondition, "このサーバーでは Google ログインが設定されていません")
	ErrSignupDisabled       = New(ErrFailedPrecondition, "このワークスペースでは新規登録を受け付けていません")

	ErrInvalidTimeZone = New(ErrValidation, "タイムゾーンの指定が正しくありません")
)
