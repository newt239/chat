package rpc

import (
	"errors"

	"connectrpc.com/connect"
	"go.uber.org/zap"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	"github.com/newt239/chat/internal/infrastructure/logger"
	adminuc "github.com/newt239/chat/internal/usecase/admin"
	appuc "github.com/newt239/chat/internal/usecase/app"
	bookmarkuc "github.com/newt239/chat/internal/usecase/bookmark"
	channeluc "github.com/newt239/chat/internal/usecase/channel"
	channelcategoryuc "github.com/newt239/chat/internal/usecase/channelcategory"
	channellinkuc "github.com/newt239/chat/internal/usecase/channellink"
	channelmemberuc "github.com/newt239/chat/internal/usecase/channelmember"
	customemojiuc "github.com/newt239/chat/internal/usecase/customemoji"
	dmuc "github.com/newt239/chat/internal/usecase/dm"
	draftuc "github.com/newt239/chat/internal/usecase/draft"
	insightuc "github.com/newt239/chat/internal/usecase/insight"
	invitationuc "github.com/newt239/chat/internal/usecase/invitation"
	mentionuc "github.com/newt239/chat/internal/usecase/mention"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	pinuc "github.com/newt239/chat/internal/usecase/pin"
	polluc "github.com/newt239/chat/internal/usecase/poll"
	reactionuc "github.com/newt239/chat/internal/usecase/reaction"
	readstateuc "github.com/newt239/chat/internal/usecase/readstate"
	scheduledmessageuc "github.com/newt239/chat/internal/usecase/scheduledmessage"
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
		domerr.ErrNotFound, domerr.ErrMessageNotFound, domerr.ErrChannelNotFound, domerr.ErrInvitationNotFound,
		adminuc.ErrMemberNotFound,
		appuc.ErrAppNotFound,
		polluc.ErrPollNotFound,
		bookmarkuc.ErrMessageNotFound,
		entity.ErrUserNotFound,
		channeluc.ErrWorkspaceNotFound, channeluc.ErrChannelNotFound,
		channelcategoryuc.ErrCategoryNotFound,
		channellinkuc.ErrLinkNotFound,
		channelmemberuc.ErrChannelNotFound, channelmemberuc.ErrUserNotFound,
		customemojiuc.ErrEmojiNotFound,
		draftuc.ErrParentMessageNotFound,
		scheduledmessageuc.ErrScheduledMessageNotFound,
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
		domerr.ErrUnauthorized, domerr.ErrForbidden, domerr.ErrNotChannelMember, domerr.ErrInvitationRequired, domerr.ErrEmailNotVerified,
		adminuc.ErrOwnerOnlyPermissions,
		appuc.ErrUnauthorized, appuc.ErrOfficialApp, appuc.ErrForbiddenChannel, appuc.ErrForbiddenThread,
		bookmarkuc.ErrUnauthorized,
		channeluc.ErrUnauthorized,
		channelcategoryuc.ErrUnauthorized,
		channellinkuc.ErrUnauthorized,
		channelmemberuc.ErrUnauthorized, channelmemberuc.ErrChannelNotPublic,
		customemojiuc.ErrUnauthorized,
		dmuc.ErrNotWorkspaceMember,
		mentionuc.ErrUnauthorized,
		messageuc.ErrUnauthorized, messageuc.ErrOfficialMessage,
		polluc.ErrUnauthorized,
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
		channeluc.ErrChannelNameExists,
		channelmemberuc.ErrAlreadyMember,
		customemojiuc.ErrNameExists,
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
		useruc.ErrInvalidTimeZone, useruc.ErrInvalidLink,
		workspaceuc.ErrInvalidRole,
	}},
	{connect.CodeFailedPrecondition, []error{
		domerr.ErrChannelArchived, domerr.ErrPasswordAuthDisabled, domerr.ErrGoogleAuthDisabled, domerr.ErrSignupDisabled,
		adminuc.ErrCannotSuspendOwner, adminuc.ErrCannotSuspendSelf,
		appuc.ErrUnsupportedChannel, appuc.ErrInactive,
		polluc.ErrPollClosed,
		channeluc.ErrCannotArchiveDM,
		channeluc.ErrChannelHasChildren,
		channelmemberuc.ErrNotMember, channelmemberuc.ErrLastAdminRemoval,
		messageuc.ErrMessageAlreadyDeleted, messageuc.ErrCannotEditDeleted,
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
