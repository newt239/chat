package rpc

import (
	"errors"

	"connectrpc.com/connect"
	"go.uber.org/zap"

	domerr "github.com/newt239/chat/internal/domain/errors"
	"github.com/newt239/chat/internal/infrastructure/logger"
	adminuc "github.com/newt239/chat/internal/usecase/admin"
	appuc "github.com/newt239/chat/internal/usecase/app"
	attachmentuc "github.com/newt239/chat/internal/usecase/attachment"
	channeluc "github.com/newt239/chat/internal/usecase/channel"
	channelcategoryuc "github.com/newt239/chat/internal/usecase/channelcategory"
	channellinkuc "github.com/newt239/chat/internal/usecase/channellink"
	channelmemberuc "github.com/newt239/chat/internal/usecase/channelmember"
	customemojiuc "github.com/newt239/chat/internal/usecase/customemoji"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	polluc "github.com/newt239/chat/internal/usecase/poll"
	scheduledmessageuc "github.com/newt239/chat/internal/usecase/scheduledmessage"
	usergroupuc "github.com/newt239/chat/internal/usecase/user_group"
	workspaceuc "github.com/newt239/chat/internal/usecase/workspace"
)

// errorCodes は ErrValidation を包んだエラーを InvalidArgument として扱うため、入力の検証エラーは個別に並べない
var errorCodes = []struct {
	code connect.Code
	errs []error
}{
	{connect.CodeNotFound, []error{
		domerr.ErrNotFound, domerr.ErrWorkspaceNotFound, domerr.ErrUserNotFound, domerr.ErrChannelNotFound,
		domerr.ErrMessageNotFound, domerr.ErrParentMessageNotFound, domerr.ErrAttachmentNotFound, domerr.ErrInvitationNotFound,
		adminuc.ErrMemberNotFound,
		appuc.ErrAppNotFound,
		attachmentuc.ErrThumbnailNotFound,
		channelcategoryuc.ErrCategoryNotFound,
		channellinkuc.ErrLinkNotFound,
		customemojiuc.ErrEmojiNotFound,
		polluc.ErrPollNotFound,
		scheduledmessageuc.ErrScheduledMessageNotFound,
		usergroupuc.ErrUserGroupNotFound,
	}},
	{connect.CodeUnauthenticated, []error{
		domerr.ErrInvalidCredentials, domerr.ErrInvalidToken, domerr.ErrSessionNotFound,
	}},
	{connect.CodePermissionDenied, []error{
		domerr.ErrUnauthorized, domerr.ErrForbidden, domerr.ErrNotChannelMember, domerr.ErrInvitationRequired, domerr.ErrEmailNotVerified,
		adminuc.ErrOwnerOnlyPermissions,
		appuc.ErrOfficialApp, appuc.ErrForbiddenChannel, appuc.ErrForbiddenThread,
		channelmemberuc.ErrChannelNotPublic,
		messageuc.ErrOfficialMessage,
		workspaceuc.ErrWorkspaceNotPublic,
	}},
	{connect.CodeAlreadyExists, []error{
		domerr.ErrUserAlreadyExists, domerr.ErrAlreadyMember, domerr.ErrPinExists, domerr.ErrReactionExists, domerr.ErrBookmarkExists, domerr.ErrWorkspaceIDExists,
		channeluc.ErrChannelNameExists,
		customemojiuc.ErrNameExists,
		usergroupuc.ErrUserGroupNameExists, usergroupuc.ErrUserAlreadyInGroup,
	}},
	{connect.CodeInvalidArgument, []error{
		domerr.ErrInvalidInput, domerr.ErrValidation,
		attachmentuc.ErrThumbnailNotAllowed,
	}},
	{connect.CodeFailedPrecondition, []error{
		domerr.ErrChannelArchived, domerr.ErrPasswordAuthDisabled, domerr.ErrGoogleAuthDisabled, domerr.ErrSignupDisabled,
		adminuc.ErrCannotSuspendOwner, adminuc.ErrCannotSuspendSelf,
		appuc.ErrUnsupportedChannel, appuc.ErrInactive,
		channeluc.ErrCannotArchiveDM, channeluc.ErrChannelHasChildren,
		channelmemberuc.ErrNotMember, channelmemberuc.ErrLastAdminRemoval,
		messageuc.ErrMessageAlreadyDeleted, messageuc.ErrCannotEditDeleted,
		polluc.ErrPollClosed,
		scheduledmessageuc.ErrNotEditable,
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
