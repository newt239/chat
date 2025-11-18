package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/newt239/chat/internal/openapi_gen"
	"github.com/newt239/chat/internal/usecase/attachment"
)

type AttachmentHandler struct {
	AttachmentUseCase *attachment.Interactor
}

type PresignUploadResponse struct {
	AttachmentID string `json:"attachmentId"`
	UploadURL    string `json:"uploadUrl"`
	StorageKey   string `json:"storageKey"`
	ExpiresAt    string `json:"expiresAt"`
}

type AttachmentMetadataResponse struct {
	ID         string  `json:"id"`
	MessageID  *string `json:"messageId,omitempty"`
	UploaderID string  `json:"uploaderId"`
	ChannelID  string  `json:"channelId"`
	FileName   string  `json:"fileName"`
	MimeType   string  `json:"mimeType"`
	SizeBytes  int64   `json:"sizeBytes"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"createdAt"`
}

type DownloadURLResponse struct {
	URL       string `json:"url"`
	ExpiresIn int    `json:"expiresIn"`
}

func (h *AttachmentHandler) PresignUpload(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "認証が必要です")
	}

	var req openapi.PresignRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "リクエストが不正です")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := &attachment.PresignInput{
		UserID:     userID,
		ChannelID:  req.ChannelId.String(),
		FileName:   req.FileName,
		MimeType:   req.ContentType,
		SizeBytes:  int64(req.SizeBytes),
		ExpiresMin: 0, // 生成型にExpiresMinフィールドがないためデフォルト値を使用
	}

	output, err := h.AttachmentUseCase.Presign(c.Request().Context(), input)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, PresignUploadResponse{
		AttachmentID: output.AttachmentID,
		UploadURL:    output.UploadURL,
		StorageKey:   output.StorageKey,
		ExpiresAt:    output.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetAttachment はServerInterfaceのGetAttachmentメソッドを実装します
func (h *AttachmentHandler) GetAttachment(ctx echo.Context, id openapi_types.UUID) error {
	userID, ok := ctx.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "認証が必要です")
	}

	output, err := h.AttachmentUseCase.GetMetadata(ctx.Request().Context(), userID, id.String())
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return ctx.JSON(http.StatusOK, AttachmentMetadataResponse{
		ID:         output.ID,
		MessageID:  output.MessageID,
		UploaderID: output.UploaderID,
		ChannelID:  output.ChannelID,
		FileName:   output.FileName,
		MimeType:   output.MimeType,
		SizeBytes:  output.SizeBytes,
		Status:     output.Status,
		CreatedAt:  output.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// DownloadAttachment はServerInterfaceのDownloadAttachmentメソッドを実装します
func (h *AttachmentHandler) DownloadAttachment(ctx echo.Context, id openapi_types.UUID) error {
	userID, ok := ctx.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "認証が必要です")
	}

	output, err := h.AttachmentUseCase.GetDownloadURL(ctx.Request().Context(), userID, id.String())
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return ctx.JSON(http.StatusOK, DownloadURLResponse{
		URL:       output.URL,
		ExpiresIn: output.ExpiresIn,
	})
}
