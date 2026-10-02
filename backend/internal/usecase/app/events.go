package app

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

// EventSender は送信 Webhook の本文を外部の URL へ送ります
type EventSender interface {
	Send(ctx context.Context, url string, body []byte, headers map[string]string) error
}

// EventDispatcher は参加しているチャンネルに投稿があったことを、送信 Webhook を許可したアプリへ知らせます
type EventDispatcher struct {
	appRepo    domainrepository.AppRepository
	mentionSvc domainservice.MentionService
	sender     EventSender
	logger     domainservice.Logger
}

func NewEventDispatcher(appRepo domainrepository.AppRepository, mentionSvc domainservice.MentionService, sender EventSender, logger domainservice.Logger) *EventDispatcher {
	return &EventDispatcher{appRepo: appRepo, mentionSvc: mentionSvc, sender: sender, logger: logger}
}

type eventPayload struct {
	Type        string       `json:"type"`
	AppID       string       `json:"app_id"`
	WorkspaceID string       `json:"workspace_id"`
	Channel     eventChannel `json:"channel"`
	Message     eventMessage `json:"message"`
}

type eventChannel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type eventMessage struct {
	ID       string  `json:"id"`
	ParentID *string `json:"parent_id"`
	// 名前に置き換えた本文と、メンションを ID で書いた元の本文
	Text      string    `json:"text"`
	RawText   string    `json:"raw_text"`
	User      eventUser `json:"user"`
	CreatedAt time.Time `json:"created_at"`
}

type eventUser struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	IsApp       bool   `json:"is_app"`
}

// NotifyNewMessage は投稿の応答を待たせないよう非同期で送り、失敗はログに残すだけにします
func (d *EventDispatcher) NotifyNewMessage(ctx context.Context, channel *entity.Channel, message messageuc.MessageOutput) {
	go func() {
		if err := d.dispatch(context.WithoutCancel(ctx), channel, message); err != nil {
			d.logger.Warn("送信 Webhook の送信に失敗しました", domainservice.LogField{Key: "messageID", Value: message.ID}, domainservice.LogField{Key: "error", Value: err.Error()})
		}
	}()
}

func (d *EventDispatcher) dispatch(ctx context.Context, channel *entity.Channel, message messageuc.MessageOutput) error {
	apps, err := d.appRepo.FindByChannelID(ctx, channel.ID)
	if err != nil {
		return fmt.Errorf("failed to load apps: %w", err)
	}
	var targets []*entity.App
	for _, a := range apps {
		// 自分の投稿を受け取って投稿し返すループを避ける
		if a.Has(entity.AppPermissionOutgoingWebhook) && a.OutgoingURL != nil && a.OutgoingSecret != nil && a.BotUserID != message.UserID {
			targets = append(targets, a)
		}
	}
	if len(targets) == 0 {
		return nil
	}
	text, err := d.mentionSvc.RenderPlain(ctx, message.Body)
	if err != nil {
		return fmt.Errorf("failed to render message body: %w", err)
	}

	var errs []error
	for _, a := range targets {
		body, err := json.Marshal(eventPayload{
			Type:        "message.created",
			AppID:       a.ID,
			WorkspaceID: channel.WorkspaceID,
			Channel:     eventChannel{ID: channel.ID, Name: channel.Name},
			Message: eventMessage{
				ID:        message.ID,
				ParentID:  message.ParentID,
				Text:      text,
				RawText:   message.Body,
				User:      eventUser{ID: message.UserID, DisplayName: message.User.DisplayName, IsApp: message.User.IsApp},
				CreatedAt: message.CreatedAt,
			},
		})
		if err != nil {
			return err
		}
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		headers := map[string]string{
			"Content-Type":     "application/json",
			"X-Chat-Timestamp": timestamp,
			"X-Chat-Signature": "sha256=" + Sign(*a.OutgoingSecret, timestamp, body),
		}
		if err := d.sender.Send(ctx, *a.OutgoingURL, body, headers); err != nil {
			errs = append(errs, fmt.Errorf("app %s: %w", a.ID, err))
		}
	}
	return errors.Join(errs...)
}

// Sign は送信 Webhook の署名です。受け取り側は「タイムスタンプ.本文」の HMAC-SHA256 を秘密鍵で計算して照合する
func Sign(secret, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "."))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
