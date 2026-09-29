package rpc

import (
	"errors"

	"connectrpc.com/connect"
	"go.uber.org/zap"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	"github.com/newt239/chat/internal/infrastructure/logger"
	adminuc "github.com/newt239/chat/internal/usecase/admin"
	bookmarkuc "github.com/newt239/chat/internal/usecase/bookmark"
	channeluc "github.com/newt239/chat/internal/usecase/channel"
	channellinkuc "github.com/newt239/chat/internal/usecase/channellink"
	channelmemberuc "github.com/newt239/chat/internal/usecase/channelmember"
	dmuc "github.com/newt239/chat/internal/usecase/dm"
	draftuc "github.com/newt239/chat/internal/usecase/draft"
	insightuc "github.com/newt239/chat/internal/usecase/insight"
	invitationuc "github.com/newt239/chat/internal/usecase/invitation"
	mentionuc "github.com/newt239/chat/internal/usecase/mention"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	pinuc "github.com/newt239/chat/internal/usecase/pin"
	reactionuc "github.com/newt239/chat/internal/usecase/reaction"
	readstateuc "github.com/newt239/chat/internal/usecase/readstate"
	scheduledmessageuc "github.com/newt239/chat/internal/usecase/scheduledmessage"
	searchuc "github.com/newt239/chat/internal/usecase/search"
	useruc "github.com/newt239/chat/internal/usecase/user"
	usergroupuc "github.com/newt239/chat/internal/usecase/user_group"
	webhookuc "github.com/newt239/chat/internal/usecase/webhook"
	workspaceuc "github.com/newt239/chat/internal/usecase/workspace"
)

var errorCodes = []struct {
	code connect.Code
	errs []error
}{
	{connect.CodeNotFound, []error{
		domerr.ErrNotFound, domerr.ErrMessageNotFound, domerr.ErrChannelNotFound, domerr.ErrInvitationNotFound,
		adminuc.ErrMemberNotFound,
		bookmarkuc.ErrMessageNotFound,
		entity.ErrUserNotFound,
		channeluc.ErrWorkspaceNotFound, channeluc.ErrChannelNotFound,
		channellinkuc.ErrLinkNotFound,
		channelmemberuc.ErrChannelNotFound, channelmemberuc.ErrUserNotFound,
		draftuc.ErrParentMessageNotFound,
		scheduledmessageuc.ErrScheduledMessageNotFound,
		messageuc.ErrChannelNotFound, messageuc.ErrParentMessageNotFound, messageuc.ErrMessageNotFound, messageuc.ErrAttachmentNotFound,
		pinuc.ErrMessageNotFound,
		reactionuc.ErrMessageNotFound,
		readstateuc.ErrChannelNotFound,
		searchuc.ErrWorkspaceNotFound,
		usergroupuc.ErrUserGroupNotFound,
		webhookuc.ErrWebhookNotFound,
		workspaceuc.ErrWorkspaceNotFound,
	}},
	{connect.CodeUnauthenticated, []error{
		domerr.ErrInvalidCredentials, domerr.ErrInvalidToken, domerr.ErrSessionNotFound,
	}},
	{connect.CodePermissionDenied, []error{
		domerr.ErrUnauthorized, domerr.ErrForbidden, domerr.ErrInvitationRequired, domerr.ErrEmailNotVerified,
		adminuc.ErrOwnerOnlyPermissions,
		bookmarkuc.ErrUnauthorized,
		channeluc.ErrUnauthorized,
		channellinkuc.ErrUnauthorized,
		channelmemberuc.ErrUnauthorized, channelmemberuc.ErrChannelNotPublic,
		dmuc.ErrNotWorkspaceMember,
		mentionuc.ErrUnauthorized,
		messageuc.ErrUnauthorized,
		pinuc.ErrUnauthorized,
		reactionuc.ErrUnauthorized,
		readstateuc.ErrUnauthorized,
		searchuc.ErrUnauthorized,
		useruc.ErrUnauthorized,
		usergroupuc.ErrUnauthorized,
		webhookuc.ErrUnauthorized,
		workspaceuc.ErrUnauthorized,
	}},
	{connect.CodeAlreadyExists, []error{
		domerr.ErrUserAlreadyExists,
		bookmarkuc.ErrBookmarkExists,
		channeluc.ErrChannelNameExists,
		channelmemberuc.ErrAlreadyMember,
		invitationuc.ErrAlreadyMember,
		pinuc.ErrPinExists,
		reactionuc.ErrReactionExists,
		usergroupuc.ErrUserGroupNameExists, usergroupuc.ErrUserAlreadyInGroup,
	}},
	{connect.CodeInvalidArgument, []error{
		domerr.ErrInvalidInput, domerr.ErrValidation,
		adminuc.ErrInvalidPermission,
		insightuc.ErrInvalidTimeZone,
		invitationuc.ErrInvalidRole,
		entity.ErrGroupDMMaxMembers,
		channeluc.ErrMemberNotInWorkspace,
		channelmemberuc.ErrInvalidRole,
		messageuc.ErrEmptyMessage,
		scheduledmessageuc.ErrScheduleInPast,
		searchuc.ErrInvalidQuery, searchuc.ErrInvalidDateRange,
		useruc.ErrInvalidTimeZone,
		workspaceuc.ErrInvalidRole,
	}},
	{connect.CodeFailedPrecondition, []error{
		domerr.ErrChannelArchived, domerr.ErrPasswordAuthDisabled, domerr.ErrGoogleAuthDisabled, domerr.ErrSignupDisabled,
		adminuc.ErrCannotSuspendOwner, adminuc.ErrCannotSuspendSelf,
		channeluc.ErrCannotArchiveDM,
		channeluc.ErrChannelHasChildren,
		channelmemberuc.ErrNotMember, channelmemberuc.ErrLastAdminRemoval,
		messageuc.ErrMessageAlreadyDeleted, messageuc.ErrCannotEditDeleted,
		scheduledmessageuc.ErrNotEditable,
		usergroupuc.ErrUserNotInGroup,
		webhookuc.ErrUnsupportedChannel,
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
