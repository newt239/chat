package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/infrastructure/utils"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

type MessageHandler struct {
	messageUC messageuc.MessageUseCase
}

// ListMessagesWithThread はスレッド情報付きのメッセージ一覧を取得します
func (h *MessageHandler) ListMessagesWithThread(c echo.Context) error {
	channelID := c.Param("channelId")
	if channelID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Channel ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	limit := 20
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	var sinceTime *time.Time
	if sinceStr := c.QueryParam("since"); sinceStr != "" {
		if t, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			sinceTime = &t
		}
	}

	var untilTime *time.Time
	if untilStr := c.QueryParam("until"); untilStr != "" {
		if t, err := time.Parse(time.RFC3339, untilStr); err == nil {
			untilTime = &t
		}
	}

	input := messageuc.ListMessagesInput{
		ChannelID: channelID,
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

// GetThreadReplies は特定のメッセージのスレッド返信一覧と親メッセージを取得します
func (h *MessageHandler) GetThreadReplies(c echo.Context) error {
	messageID, err := utils.GetParamFromContext(c, "messageId")
	if err != nil {
		return err
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	input := messageuc.GetThreadRepliesInput{
		MessageID: messageID,
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

// GetThreadMetadata は特定のメッセージのスレッドメタデータを取得します
func (h *MessageHandler) GetThreadMetadata(c echo.Context) error {
	messageID, err := utils.GetParamFromContext(c, "messageId")
	if err != nil {
		return err
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	input := messageuc.GetThreadMetadataInput{
		MessageID: messageID,
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

// CreateMessageRequest はメッセージ作成リクエストの構造体です
type CreateMessageRequest struct {
	Body     string  `json:"body" validate:"required,min=1"`
	ParentID *string `json:"parentId,omitempty"`
}

// UpdateMessageRequest はメッセージ更新リクエストの構造体です
type UpdateMessageRequest struct {
	Body string `json:"body" validate:"required,min=1"`
}

// ListMessages はメッセージ一覧を取得します
func (h *MessageHandler) ListMessages(c echo.Context) error {
	channelID := c.Param("channelId")
	if channelID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Channel ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	limitStr := c.QueryParam("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// 日付フィルター
	sinceStr := c.QueryParam("since")
	since := ""
	if sinceStr != "" {
		since = sinceStr
	}

	untilStr := c.QueryParam("until")
	until := ""
	if untilStr != "" {
		until = untilStr
	}

	var sinceTime *time.Time
	if since != "" {
		if t, err := time.Parse(time.RFC3339, since); err == nil {
			sinceTime = &t
		}
	}

	var untilTime *time.Time
	if until != "" {
		if t, err := time.Parse(time.RFC3339, until); err == nil {
			untilTime = &t
		}
	}

	input := messageuc.ListMessagesInput{
		ChannelID: channelID,
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

// CreateMessage はメッセージを作成します
func (h *MessageHandler) CreateMessage(c echo.Context) error {
	channelID, err := utils.GetParamFromContext(c, "channelId")
	if err != nil {
		return err
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	var req CreateMessageRequest
	if err := utils.ValidateRequest(c, &req); err != nil {
		return err
	}

	input := messageuc.CreateMessageInput{
		ChannelID: channelID,
		UserID:    userID,
		Body:      req.Body,
		ParentID:  req.ParentID,
	}

	message, err := h.messageUC.CreateMessage(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusCreated, message)
}

// UpdateMessage はメッセージを更新します
func (h *MessageHandler) UpdateMessage(c echo.Context) error {
	messageID, err := utils.GetParamFromContext(c, "messageId")
	if err != nil {
		return err
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	var req UpdateMessageRequest
	if err := utils.ValidateRequest(c, &req); err != nil {
		return err
	}

	input := messageuc.UpdateMessageInput{
		MessageID: messageID,
		EditorID:  userID,
		Body:      req.Body,
	}

	message, err := h.messageUC.UpdateMessage(c.Request().Context(), input)
	if err != nil {
		return mapMessageError(err)
	}

	return c.JSON(http.StatusOK, message)
}

// DeleteMessage はメッセージを削除します
func (h *MessageHandler) DeleteMessage(c echo.Context) error {
	messageID, err := utils.GetParamFromContext(c, "messageId")
	if err != nil {
		return err
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	input := messageuc.DeleteMessageInput{
		MessageID:  messageID,
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
