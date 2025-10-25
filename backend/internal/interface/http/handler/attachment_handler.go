package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/example/chat/internal/usecase/attachment"
)

type AttachmentHandler struct {
	attachmentUseCase *attachment.Interactor
}

func NewAttachmentHandler(attachmentUseCase *attachment.Interactor) *AttachmentHandler {
	return &AttachmentHandler{
		attachmentUseCase: attachmentUseCase,
	}
}

type PresignUploadRequest struct {
	ChannelID  string `json:"channelId" validate:"required"`
	FileName   string `json:"fileName" validate:"required"`
	MimeType   string `json:"mimeType" validate:"required"`
	SizeBytes  int64  `json:"sizeBytes" validate:"required,min=1"`
	ExpiresMin int    `json:"expiresMin,omitempty"`
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

	var req PresignUploadRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "リクエストが不正です")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := &attachment.PresignInput{
		UserID:     userID,
		ChannelID:  req.ChannelID,
		FileName:   req.FileName,
		MimeType:   req.MimeType,
		SizeBytes:  req.SizeBytes,
		ExpiresMin: req.ExpiresMin,
	}

	output, err := h.attachmentUseCase.Presign(c.Request().Context(), input)
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

func (h *AttachmentHandler) GetMetadata(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "認証が必要です")
	}

	attachmentID := c.Param("attachmentId")
	if attachmentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "添付ファイルIDが必要です")
	}

	output, err := h.attachmentUseCase.GetMetadata(c.Request().Context(), userID, attachmentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, AttachmentMetadataResponse{
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

func (h *AttachmentHandler) GetDownloadURL(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "認証が必要です")
	}

	attachmentID := c.Param("attachmentId")
	if attachmentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "添付ファイルIDが必要です")
	}

	output, err := h.attachmentUseCase.GetDownloadURL(c.Request().Context(), userID, attachmentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, DownloadURLResponse{
		URL:       output.URL,
		ExpiresIn: output.ExpiresIn,
	})
}
