package handler

import (
	"errors"
	"net/http"
	"strings"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/labstack/echo/v4"

	"github.com/newt239/chat/internal/usecase/channelmember"
	"github.com/newt239/chat/internal/usecase/systemmessage"
	"github.com/newt239/chat/internal/domain/entity"
)

type ChannelMemberHandler struct {
	channelMemberUseCase channelmember.ChannelMemberUseCase
    systemMessageUC      systemmessage.UseCase
}

func NewChannelMemberHandler(channelMemberUseCase channelmember.ChannelMemberUseCase, systemMessageUC systemmessage.UseCase) *ChannelMemberHandler {
    return &ChannelMemberHandler{channelMemberUseCase: channelMemberUseCase, systemMessageUC: systemMessageUC}
}

type InviteMemberRequest struct {
	UserID string  `json:"userId"`
	Role   *string `json:"role"`
}

func (r *InviteMemberRequest) Validate() error {
	if strings.TrimSpace(r.UserID) == "" {
		return errors.New("userIdは必須です")
	}
	if r.Role != nil && (*r.Role != "member" && *r.Role != "admin") {
		return errors.New("roleは'member'または'admin'を指定してください")
	}
	return nil
}

type ChannelMemberUpdateRoleRequest struct {
	Role string `json:"role"`
}

func (r *ChannelMemberUpdateRoleRequest) Validate() error {
	if strings.TrimSpace(r.Role) == "" {
		return errors.New("roleは必須です")
	}
	return nil
}

type SuccessResponse struct {
	Success bool `json:"success"`
}

// ListChannelMembers はチャンネルメンバー一覧を取得します (ServerInterface用)
func (h *ChannelMemberHandler) ListChannelMembers(c echo.Context, channelId openapi_types.UUID) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "ユーザーが認証されていません"})
	}

	output, err := h.channelMemberUseCase.ListMembers(c.Request().Context(), channelmember.ListMembersInput{
		ChannelID: channelId.String(),
		UserID:    userID,
	})
	if err != nil {
		switch err {
		case channelmember.ErrChannelNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "メンバー一覧の取得に失敗しました"})
		}
	}

	return c.JSON(http.StatusOK, output)
}

// InviteChannelMember はチャンネルにメンバーを招待します (ServerInterface用)
func (h *ChannelMemberHandler) InviteChannelMember(c echo.Context, channelId openapi_types.UUID) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "ユーザーが認証されていません"})
	}

	var req InviteMemberRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "リクエストボディが不正です: " + err.Error()})
	}

	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	role := "member"
	if req.Role != nil {
		role = *req.Role
	}

    err := h.channelMemberUseCase.InviteMember(c.Request().Context(), channelmember.InviteMemberInput{
		ChannelID:    channelId.String(),
		OperatorID:   userID,
		TargetUserID: req.UserID,
		Role:         role,
	})
	if err != nil {
		switch err {
		case channelmember.ErrChannelNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrUserNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case channelmember.ErrAlreadyMember:
			return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		case channelmember.ErrInvalidRole:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "メンバー招待に失敗しました"})
		}
	}

    // システムメッセージ: member_added
    if h.systemMessageUC != nil {
        actorID := userID
        payload := map[string]any{"userId": req.UserID, "addedBy": userID}
        _, _ = h.systemMessageUC.Create(c.Request().Context(), systemmessage.CreateInput{
            ChannelID: channelId.String(),
            Kind:      entity.SystemMessageKindMemberAdded,
            Payload:   payload,
            ActorID:   &actorID,
        })
    }

	return c.JSON(http.StatusOK, SuccessResponse{Success: true})
}

// JoinPublicChannel はパブリックチャンネルに参加します (ServerInterface用)
func (h *ChannelMemberHandler) JoinPublicChannel(c echo.Context, channelId openapi_types.UUID) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "ユーザーが認証されていません"})
	}

    err := h.channelMemberUseCase.JoinPublicChannel(c.Request().Context(), channelmember.JoinChannelInput{
		ChannelID: channelId.String(),
		UserID:    userID,
	})
	if err != nil {
		switch err {
		case channelmember.ErrChannelNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrChannelNotPublic:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case channelmember.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "チャンネルへの参加に失敗しました"})
		}
	}

    // システムメッセージ: member_joined
    if h.systemMessageUC != nil {
        actorID := userID
        payload := map[string]any{"userId": userID}
        _, _ = h.systemMessageUC.Create(c.Request().Context(), systemmessage.CreateInput{
            ChannelID: channelId.String(),
            Kind:      entity.SystemMessageKindMemberJoined,
            Payload:   payload,
            ActorID:   &actorID,
        })
    }

	return c.JSON(http.StatusOK, SuccessResponse{Success: true})
}

// UpdateChannelMemberRole はチャンネルメンバーの権限を更新します (ServerInterface用)
func (h *ChannelMemberHandler) UpdateChannelMemberRole(c echo.Context, channelId openapi_types.UUID, userId openapi_types.UUID) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "ユーザーが認証されていません"})
	}

	var req ChannelMemberUpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "リクエストボディが不正です: " + err.Error()})
	}

	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	err := h.channelMemberUseCase.UpdateMemberRole(c.Request().Context(), channelmember.UpdateMemberRoleInput{
		ChannelID:    channelId.String(),
		OperatorID:   userID,
		TargetUserID: userId.String(),
		Role:         req.Role,
	})
	if err != nil {
		switch err {
		case channelmember.ErrChannelNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrNotMember:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case channelmember.ErrInvalidRole:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		case channelmember.ErrLastAdminRemoval:
			return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "メンバー権限の更新に失敗しました"})
		}
	}

	return c.JSON(http.StatusOK, SuccessResponse{Success: true})
}

// RemoveChannelMember はチャンネルからメンバーを削除します (ServerInterface用)
func (h *ChannelMemberHandler) RemoveChannelMember(c echo.Context, channelId openapi_types.UUID, userId openapi_types.UUID) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "ユーザーが認証されていません"})
	}

	err := h.channelMemberUseCase.RemoveMember(c.Request().Context(), channelmember.RemoveMemberInput{
		ChannelID:    channelId.String(),
		OperatorID:   userID,
		TargetUserID: userId.String(),
	})
	if err != nil {
		switch err {
		case channelmember.ErrChannelNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrNotMember:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case channelmember.ErrLastAdminRemoval:
			return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "メンバーの削除に失敗しました"})
		}
	}

	return c.JSON(http.StatusOK, SuccessResponse{Success: true})
}

// LeaveChannel はチャンネルから退出します (ServerInterface用)
func (h *ChannelMemberHandler) LeaveChannel(c echo.Context, channelId openapi_types.UUID) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "ユーザーが認証されていません"})
	}

	err := h.channelMemberUseCase.LeaveChannel(c.Request().Context(), channelmember.LeaveChannelInput{
		ChannelID: channelId.String(),
		UserID:    userID,
	})
	if err != nil {
		switch err {
		case channelmember.ErrChannelNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrNotMember:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case channelmember.ErrLastAdminRemoval:
			return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "チャンネルからの退出に失敗しました"})
		}
	}

	return c.JSON(http.StatusOK, SuccessResponse{Success: true})
}
