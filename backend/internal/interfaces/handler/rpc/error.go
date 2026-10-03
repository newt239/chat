package rpc

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"

	domerr "github.com/newt239/chat/internal/domain/errors"
)

// errorCodes はユースケースのエラーを種類ごとに Connect のエラーコードへ対応させます
var errorCodes = []struct {
	kind error
	code connect.Code
}{
	{domerr.ErrNotFound, connect.CodeNotFound},
	{domerr.ErrUnauthenticated, connect.CodeUnauthenticated},
	{domerr.ErrUnauthorized, connect.CodePermissionDenied},
	{domerr.ErrAlreadyExists, connect.CodeAlreadyExists},
	{domerr.ErrValidation, connect.CodeInvalidArgument},
	{domerr.ErrFailedPrecondition, connect.CodeFailedPrecondition},
	{domerr.ErrConflict, connect.CodeAborted},
}

func toConnectError(ctx context.Context, procedure string, err error) *connect.Error {
	for _, entry := range errorCodes {
		if errors.Is(err, entry.kind) {
			slog.DebugContext(ctx, "RPC がエラーを返しました", "procedure", procedure, "error", err)
			return connect.NewError(entry.code, err)
		}
	}
	// 内部エラーの詳細はクライアントに返さない
	slog.ErrorContext(ctx, "RPC で予期しないエラーが発生しました", "procedure", procedure, "error", err)
	return connect.NewError(connect.CodeInternal, errors.New("サーバー内部でエラーが発生しました"))
}
