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
	appuc "github.com/newt239/chat/internal/usecase/app"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

const maxWebhookPayloadBytes = 64 << 10

// WebhookPoster はアプリの着信 Webhook の投稿を受け付けるユースケースです
type WebhookPoster interface {
	Post(ctx context.Context, input appuc.PostInput) (*messageuc.MessageOutput, error)
}

// webhookPayload は Slack の Incoming Webhook と互換の最小限の形式です。avatar_url は Discord 互換の別名。
// channel_id を省略したらアプリの既定のチャンネルに、thread_id を指定したらそのスレッドに投稿する
type webhookPayload struct {
	Text      string  `json:"text"`
	Username  *string `json:"username"`
	IconURL   *string `json:"icon_url"`
	AvatarURL *string `json:"avatar_url"`
	ChannelID *string `json:"channel_id"`
	ThreadID  *string `json:"thread_id"`
}

var webhookErrorStatuses = []struct {
	status int
	errs   []error
}{
	{http.StatusNotFound, []error{appuc.ErrAppNotFound, domerr.ErrChannelNotFound, messageuc.ErrParentMessageNotFound}},
	{http.StatusBadRequest, []error{domerr.ErrValidation}},
	{http.StatusForbidden, []error{appuc.ErrInactive, appuc.ErrForbiddenChannel, appuc.ErrForbiddenThread, domerr.ErrChannelArchived}},
}

func webhookHandler(poster WebhookPoster, limiter RateLimiter) echo.HandlerFunc {
	return func(c echo.Context) error {
		appID := c.Param("id")
		if ok, wait := limiter.Allow(c.Request().Context(), appID); !ok {
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

		_, err = poster.Post(c.Request().Context(), appuc.PostInput{
			AppID:     appID,
			Token:     c.Param("token"),
			Text:      payload.Text,
			Username:  payload.Username,
			AvatarURL: avatarURL,
			ChannelID: payload.ChannelID,
			ParentID:  payload.ThreadID,
		})
		if err != nil {
			for _, entry := range webhookErrorStatuses {
				for _, target := range entry.errs {
					if errors.Is(err, target) {
						return c.String(entry.status, err.Error())
					}
				}
			}
			logger.Get().Error("Webhook の投稿に失敗しました", zap.String("appId", appID), zap.Error(err))
			return c.String(http.StatusInternalServerError, "internal_error")
		}
		return c.String(http.StatusOK, "ok")
	}
}
