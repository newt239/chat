package app

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

type channelAppRepo struct {
	domainrepository.AppRepository
	apps []*entity.App
}

func (r channelAppRepo) FindByChannelID(context.Context, string) ([]*entity.App, error) {
	return r.apps, nil
}

type stubMentionService struct {
	domainservice.MentionService
}

func (stubMentionService) RenderPlain(_ context.Context, body string) (string, error) {
	return strings.ReplaceAll(body, "<@u1>", "@bob"), nil
}

type sent struct {
	url     string
	body    []byte
	headers map[string]string
}

type recordingSender struct {
	sent []sent
}

func (s *recordingSender) Send(_ context.Context, url string, body []byte, headers map[string]string) error {
	s.sent = append(s.sent, sent{url: url, body: body, headers: headers})
	return nil
}

func TestDispatchSendsSignedEventsToSubscribedApps(t *testing.T) {
	url, secret := "https://example.com/hook", "secret"
	subscribed := &entity.App{ID: "a1", BotUserID: "bot-1", Permissions: []entity.AppPermission{entity.AppPermissionOutgoingWebhook}, OutgoingURL: &url, OutgoingSecret: &secret}
	notAllowed := &entity.App{ID: "a2", BotUserID: "bot-2", OutgoingURL: &url, OutgoingSecret: &secret}
	sender := &recordingSender{}
	d := NewEventDispatcher(channelAppRepo{apps: []*entity.App{subscribed, notAllowed}}, stubMentionService{}, sender)
	channel := &entity.Channel{ID: "c1", WorkspaceID: "ws", Name: "general"}

	if err := d.NotifyNewMessage(context.Background(), channel, messageuc.MessageOutput{ID: "m1", UserID: "u2", Body: "hi <@u1>"}); err != nil {
		t.Fatal(err)
	}
	if len(sender.sent) != 1 || sender.sent[0].url != url {
		t.Fatalf("送信 Webhook を許可したアプリだけに送られていません: %+v", sender.sent)
	}
	got := sender.sent[0]
	if got.headers["X-Chat-Signature"] != "sha256="+Sign(secret, got.headers["X-Chat-Timestamp"], got.body) {
		t.Fatal("署名が本文と一致しません")
	}
	var payload eventPayload
	if err := json.Unmarshal(got.body, &payload); err != nil || payload.Message.Text != "hi @bob" || payload.Message.RawText != "hi <@u1>" {
		t.Fatalf("本文が期待と異なります: %v %+v", err, payload)
	}

	// アプリ自身の投稿は送らない
	sender.sent = nil
	if err := d.NotifyNewMessage(context.Background(), channel, messageuc.MessageOutput{ID: "m2", UserID: "bot-1", Body: "echo"}); err != nil || len(sender.sent) != 0 {
		t.Fatalf("自分の投稿を送っています: %v %d", err, len(sender.sent))
	}
}
