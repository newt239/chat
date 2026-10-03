package rpc

import (
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/validate"

	authuc "github.com/newt239/chat/internal/usecase/auth"
)

// maxRequestBytes は 1 リクエストの本文の上限。ファイルは署名付き URL で直接アップロードするため小さくてよい
const maxRequestBytes = 1 << 20

// Registration は生成された NewXxxServiceHandler にサービス実装を束縛したものです
type Registration func(opts ...connect.HandlerOption) (string, http.Handler)

// Register は生成されたハンドラ生成関数とサービス実装から Registration を作ります
func Register[T any](newHandler func(T, ...connect.HandlerOption) (string, http.Handler), service T) Registration {
	return func(opts ...connect.HandlerOption) (string, http.Handler) {
		return newHandler(service, opts...)
	}
}

// NewHandler は全サービスを共通の interceptor 付きで登録した http.Handler を返します
func NewHandler(jwtService authuc.JWTService, allowedOrigins []string, registrations ...Registration) http.Handler {
	// 外側から順にエラー変換・オリジン確認・操作元の記録・認証・入力検証を適用する
	interceptors := connect.WithInterceptors(
		newErrorInterceptor(),
		newOriginInterceptor(allowedOrigins),
		newClientInfoInterceptor(),
		newAuthInterceptor(jwtService),
		validate.NewInterceptor(),
	)
	// Connect-Protocol-Version ヘッダーを必須にし、フォームなどから単純リクエストで呼ばれる CSRF を防ぐ
	opts := []connect.HandlerOption{interceptors, connect.WithRequireConnectProtocolHeader(), connect.WithReadMaxBytes(maxRequestBytes)}
	mux := http.NewServeMux()
	for _, register := range registrations {
		mux.Handle(register(opts...))
	}
	return mux
}
