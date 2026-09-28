package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	domerr "github.com/newt239/chat/internal/domain/errors"
	"github.com/newt239/chat/internal/infrastructure/logger"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	webhookuc "github.com/newt239/chat/internal/usecase/webhook"
)

const maxWebhookPayloadBytes = 64 << 10

// WebhookPoster は着信 Webhook の投稿を受け付けるユースケースです
type WebhookPoster interface {
	Post(ctx context.Context, input webhookuc.PostInput) (*messageuc.MessageOutput, error)
}

// webhookPayload は Slack の Incoming Webhook と互換の最小限の形式です。avatar_url は Discord 互換の別名
type webhookPayload struct {
	Text      string  `json:"text"`
	Username  *string `json:"username"`
	IconURL   *string `json:"icon_url"`
	AvatarURL *string `json:"avatar_url"`
}

var webhookErrorStatuses = []struct {
	status int
	errs   []error
}{
	{http.StatusNotFound, []error{webhookuc.ErrWebhookNotFound, domerr.ErrChannelNotFound}},
	{http.StatusBadRequest, []error{domerr.ErrValidation}},
	{http.StatusForbidden, []error{webhookuc.ErrInactive, domerr.ErrChannelArchived}},
}

func webhookHandler(poster WebhookPoster, limiter *rateLimiter) echo.HandlerFunc {
	return func(c echo.Context) error {
		webhookID := c.Param("id")
		if ok, wait := limiter.allow(webhookID); !ok {
			c.Response().Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(wait.Seconds()))))
			return c.String(http.StatusTooManyRequests, "rate_limited")
		}

		body, err := io.ReadAll(io.LimitReader(c.Request().Body, maxWebhookPayloadBytes+1))
		if err != nil {
			return c.String(http.StatusBadRequest, "invalid_payload")
		}
		if len(body) > maxWebhookPayloadBytes {
			return c.String(http.StatusRequestEntityTooLarge, "payload_too_large")
		}
		// curl の -d のように Content-Type が JSON でなくても本文を JSON として読む
		var payload webhookPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			return c.String(http.StatusBadRequest, "invalid_payload")
		}
		avatarURL := payload.AvatarURL
		if avatarURL == nil {
			avatarURL = payload.IconURL
		}

		_, err = poster.Post(c.Request().Context(), webhookuc.PostInput{
			WebhookID: webhookID,
			Token:     c.Param("token"),
			Text:      payload.Text,
			Username:  payload.Username,
			AvatarURL: avatarURL,
		})
		if err != nil {
			for _, entry := range webhookErrorStatuses {
				for _, target := range entry.errs {
					if errors.Is(err, target) {
						return c.String(entry.status, err.Error())
					}
				}
			}
			logger.Get().Error("Webhook の投稿に失敗しました", zap.String("webhookId", webhookID), zap.Error(err))
			return c.String(http.StatusInternalServerError, "internal_error")
		}
		return c.String(http.StatusOK, "ok")
	}
}
