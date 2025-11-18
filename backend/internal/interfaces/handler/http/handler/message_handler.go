package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/infrastructure/utils"
	openapi "github.com/newt239/chat/internal/openapi_gen"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type MessageHandler struct {
	messageUC messageuc.MessageUseCase
}

// ListMessagesWithThread はスレッド情報付きのメッセージ一覧を取得します (ServerInterface用)
func (h *MessageHandler) ListMessagesWithThread(c echo.Context, channelId openapi_types.UUID, params openapi.ListMessagesWithThreadParams) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}

	var sinceTime *time.Time
	if params.Since != nil {
		sinceTime = params.Since
	}

	var untilTime *time.Time
	if params.Until != nil {
		untilTime = params.Until
	}

	input := messageuc.ListMessagesInput{
		ChannelID: channelId.String(),
		UserID:    userID,
		Limit:     limit,
		Since:     sinceTime,
		Until:     untilTime,
	}

	// hasMore を取得するため通常の一覧も取得
	listRes, err := h.messageUC.ListMessages(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	// スレッド情報付きの一覧を取得
	outputs, err := h.messageUC.ListMessagesWithThread(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"messages": outputs,
		"hasMore":  listRes.HasMore,
	})
}

func NewMessageHandler(messageUC messageuc.MessageUseCase) *MessageHandler {
	return &MessageHandler{messageUC: messageUC}
}

// GetThreadReplies は特定のメッセージのスレッド返信一覧と親メッセージを取得します (ServerInterface用)
func (h *MessageHandler) GetThreadReplies(c echo.Context, messageId openapi_types.UUID, params openapi.GetThreadRepliesParams) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	input := messageuc.GetThreadRepliesInput{
		MessageID: messageId.String(),
		UserID:    userID,
	}

	output, err := h.messageUC.GetThreadReplies(c.Request().Context(), input)
	if err != nil {
		switch err {
		case messageuc.ErrParentMessageNotFound:
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		case messageuc.ErrUnauthorized:
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		default:
			return handleUseCaseError(err)
		}
	}

	return c.JSON(http.StatusOK, output)
}

// GetThreadMetadata は特定のメッセージのスレッドメタデータを取得します (ServerInterface用)
func (h *MessageHandler) GetThreadMetadata(c echo.Context, messageId openapi_types.UUID) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	input := messageuc.GetThreadMetadataInput{
		MessageID: messageId.String(),
		UserID:    userID,
	}

	output, err := h.messageUC.GetThreadMetadata(c.Request().Context(), input)
	if err != nil {
		switch err {
		case messageuc.ErrParentMessageNotFound:
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		case messageuc.ErrUnauthorized:
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		default:
			return handleUseCaseError(err)
		}
	}

	return c.JSON(http.StatusOK, output)
}

// ListMessages はメッセージ一覧を取得します (ServerInterface用)
func (h *MessageHandler) ListMessages(c echo.Context, channelId openapi_types.UUID, params openapi.ListMessagesParams) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}

	var sinceTime *time.Time
	if params.Since != nil {
		sinceTime = params.Since
	}

	var untilTime *time.Time
	if params.Until != nil {
		untilTime = params.Until
	}

	input := messageuc.ListMessagesInput{
		ChannelID: channelId.String(),
		UserID:    userID,
		Limit:     limit,
		Since:     sinceTime,
		Until:     untilTime,
	}

	messages, err := h.messageUC.ListMessages(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, messages)
}

// CreateMessage はメッセージを作成します (ServerInterface用)
func (h *MessageHandler) CreateMessage(c echo.Context, channelId openapi_types.UUID) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	var req openapi.CreateMessageRequest
	if err := utils.ValidateRequest(c, &req); err != nil {
		return err
	}

	var parentID *string
	if req.ParentId != nil {
		parentIDStr := req.ParentId.String()
		parentID = &parentIDStr
	}

	var attachmentIDs []string
	if req.AttachmentIds != nil {
		attachmentIDs = make([]string, len(*req.AttachmentIds))
		for i, id := range *req.AttachmentIds {
			attachmentIDs[i] = id.String()
		}
	}

	input := messageuc.CreateMessageInput{
		ChannelID:     channelId.String(),
		UserID:        userID,
		Body:          req.Body,
		ParentID:      parentID,
		AttachmentIDs: attachmentIDs,
	}

	message, err := h.messageUC.CreateMessage(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusCreated, message)
}

// UpdateMessage はメッセージを更新します (ServerInterface用)
func (h *MessageHandler) UpdateMessage(c echo.Context, messageId openapi_types.UUID) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	var req openapi.UpdateMessageRequest
	if err := utils.ValidateRequest(c, &req); err != nil {
		return err
	}

	input := messageuc.UpdateMessageInput{
		MessageID: messageId.String(),
		EditorID:  userID,
		Body:      req.Body,
	}

	message, err := h.messageUC.UpdateMessage(c.Request().Context(), input)
	if err != nil {
		return mapMessageError(err)
	}

	return c.JSON(http.StatusOK, message)
}

// DeleteMessage はメッセージを削除します (ServerInterface用)
func (h *MessageHandler) DeleteMessage(c echo.Context, messageId openapi_types.UUID) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	input := messageuc.DeleteMessageInput{
		MessageID:  messageId.String(),
		ExecutorID: userID,
	}

	if err := h.messageUC.DeleteMessage(c.Request().Context(), input); err != nil {
		return mapMessageError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func mapMessageError(err error) error {
	switch err {
	case messageuc.ErrMessageNotFound, messageuc.ErrChannelNotFound:
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case messageuc.ErrUnauthorized:
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	case messageuc.ErrMessageAlreadyDeleted, messageuc.ErrCannotEditDeleted:
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	default:
		return handleUseCaseError(err)
	}
}
