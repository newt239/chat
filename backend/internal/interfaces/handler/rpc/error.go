package rpc

import (
	"errors"

	"connectrpc.com/connect"
	"go.uber.org/zap"

	domerr "github.com/newt239/chat/internal/domain/errors"
	"github.com/newt239/chat/internal/infrastructure/logger"
	bookmarkuc "github.com/newt239/chat/internal/usecase/bookmark"
	channeluc "github.com/newt239/chat/internal/usecase/channel"
	channelmemberuc "github.com/newt239/chat/internal/usecase/channelmember"
	dmuc "github.com/newt239/chat/internal/usecase/dm"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	pinuc "github.com/newt239/chat/internal/usecase/pin"
	reactionuc "github.com/newt239/chat/internal/usecase/reaction"
	readstateuc "github.com/newt239/chat/internal/usecase/readstate"
	searchuc "github.com/newt239/chat/internal/usecase/search"
	useruc "github.com/newt239/chat/internal/usecase/user"
	usergroupuc "github.com/newt239/chat/internal/usecase/user_group"
	workspaceuc "github.com/newt239/chat/internal/usecase/workspace"
)

var errorCodes = []struct {
	code connect.Code
	errs []error
}{
	{connect.CodeNotFound, []error{
		domerr.ErrNotFound, domerr.ErrMessageNotFound, domerr.ErrChannelNotFound,
		bookmarkuc.ErrMessageNotFound,
		channeluc.ErrWorkspaceNotFound, channeluc.ErrChannelNotFound,
		channelmemberuc.ErrChannelNotFound, channelmemberuc.ErrUserNotFound,
		messageuc.ErrChannelNotFound, messageuc.ErrParentMessageNotFound, messageuc.ErrMessageNotFound, messageuc.ErrAttachmentNotFound,
		pinuc.ErrMessageNotFound,
		reactionuc.ErrMessageNotFound,
		readstateuc.ErrChannelNotFound,
		searchuc.ErrWorkspaceNotFound,
		usergroupuc.ErrUserGroupNotFound,
		workspaceuc.ErrWorkspaceNotFound,
	}},
	{connect.CodeUnauthenticated, []error{
		domerr.ErrInvalidCredentials, domerr.ErrInvalidToken, domerr.ErrSessionNotFound,
	}},
	{connect.CodePermissionDenied, []error{
		domerr.ErrUnauthorized, domerr.ErrForbidden,
		bookmarkuc.ErrUnauthorized,
		channeluc.ErrUnauthorized,
		channelmemberuc.ErrUnauthorized, channelmemberuc.ErrChannelNotPublic,
		dmuc.ErrNotWorkspaceMember,
		messageuc.ErrUnauthorized,
		pinuc.ErrUnauthorized,
		reactionuc.ErrUnauthorized,
		readstateuc.ErrUnauthorized,
		searchuc.ErrUnauthorized,
		useruc.ErrUnauthorized,
		usergroupuc.ErrUnauthorized,
		workspaceuc.ErrUnauthorized,
	}},
	{connect.CodeAlreadyExists, []error{
		domerr.ErrUserAlreadyExists,
		bookmarkuc.ErrBookmarkExists,
		channelmemberuc.ErrAlreadyMember,
		pinuc.ErrPinExists,
		reactionuc.ErrReactionExists,
		usergroupuc.ErrUserGroupNameExists, usergroupuc.ErrUserAlreadyInGroup,
	}},
	{connect.CodeInvalidArgument, []error{
		domerr.ErrInvalidInput, domerr.ErrValidation,
		channelmemberuc.ErrInvalidRole,
		searchuc.ErrInvalidQuery,
		workspaceuc.ErrInvalidRole,
	}},
	{connect.CodeFailedPrecondition, []error{
		channelmemberuc.ErrNotMember, channelmemberuc.ErrLastAdminRemoval,
		messageuc.ErrMessageAlreadyDeleted, messageuc.ErrCannotEditDeleted,
		usergroupuc.ErrUserNotInGroup,
		workspaceuc.ErrCannotRemoveOwner, workspaceuc.ErrCannotChangeOwnerRole,
	}},
	{connect.CodeAborted, []error{domerr.ErrConflict}},
}

// toConnectError はユースケースのエラーを対応する Connect のエラーコードに変換します
func toConnectError(procedure string, err error) *connect.Error {
	for _, entry := range errorCodes {
		for _, target := range entry.errs {
			if errors.Is(err, target) {
				logger.Get().Debug("RPC がエラーを返しました", zap.String("procedure", procedure), zap.Error(err))
				return connect.NewError(entry.code, err)
			}
		}
	}
	// 内部エラーの詳細はクライアントに返さない
	logger.Get().Error("RPC で予期しないエラーが発生しました", zap.String("procedure", procedure), zap.Error(err))
	return connect.NewError(connect.CodeInternal, errors.New("サーバー内部でエラーが発生しました"))
}
