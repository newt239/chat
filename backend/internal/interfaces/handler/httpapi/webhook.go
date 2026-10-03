package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	"github.com/newt239/chat/internal/infrastructure/redis"
	appuc "github.com/newt239/chat/internal/usecase/app"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

const maxWebhookPayloadBytes = 64 << 10

// WebhookPoster はアプリの着信 Webhook の投稿を受け付けるユースケースです
type WebhookPoster interface {
	Authenticate(ctx context.Context, appID, token string) (*entity.App, error)
	Post(ctx context.Context, app *entity.App, input appuc.PostInput) (*messageuc.MessageOutput, error)
}

// webhookPayload は Slack の Incoming Webhook 互換の形式。channel_id を省略したら既定のチャンネル、thread_id があればスレッドに投稿する
type webhookPayload struct {
	Text      string  `json:"text"`
	ChannelID *string `json:"channel_id"`
	ThreadID  *string `json:"thread_id"`
}

var webhookErrorStatuses = []struct {
	kind   error
	status int
}{
	{domerr.ErrNotFound, http.StatusNotFound},
	{domerr.ErrValidation, http.StatusBadRequest},
	{domerr.ErrUnauthorized, http.StatusForbidden},
	{domerr.ErrFailedPrecondition, http.StatusForbidden},
}

// Webhook ごとに毎秒 1 回、瞬間的には 10 回まで受け付ける
const (
	WebhookRatePerSecond = 1
	WebhookBurst         = 10
)

func webhookHandler(poster WebhookPoster, limiter *redis.RateLimiter) echo.HandlerFunc {
	return func(c echo.Context) error {
		appID := c.Param("id")
		// 不正なトークンで他のアプリの枠を使い切られないよう、確かめてから数える
		app, err := poster.Authenticate(c.Request().Context(), appID, c.Param("token"))
		if err != nil {
			return webhookError(c, appID, err)
		}
		if ok, wait := limiter.Allow(c.Request().Context(), app.ID); !ok {
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
		if _, err := poster.Post(c.Request().Context(), app, appuc.PostInput{Text: payload.Text, ChannelID: payload.ChannelID, ParentID: payload.ThreadID}); err != nil {
			return webhookError(c, appID, err)
		}
		return c.String(http.StatusOK, "ok")
	}
}

func webhookError(c echo.Context, appID string, err error) error {
	for _, entry := range webhookErrorStatuses {
		if errors.Is(err, entry.kind) {
			return c.String(entry.status, err.Error())
		}
	}
	slog.ErrorContext(c.Request().Context(), "Webhook の投稿に失敗しました", "appId", appID, "error", err)
	return c.String(http.StatusInternalServerError, "internal_error")
}
