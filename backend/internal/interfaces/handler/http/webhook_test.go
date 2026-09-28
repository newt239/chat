package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	domerr "github.com/newt239/chat/internal/domain/errors"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	webhookuc "github.com/newt239/chat/internal/usecase/webhook"
)

type stubPoster struct {
	got webhookuc.PostInput
	err error
}

func (p *stubPoster) Post(_ context.Context, input webhookuc.PostInput) (*messageuc.MessageOutput, error) {
	p.got = input
	return &messageuc.MessageOutput{}, p.err
}

func postWebhook(t *testing.T, poster *stubPoster, body string) *httptest.ResponseRecorder {
	t.Helper()
	e := NewRouter(RouterConfig{WebhookPoster: poster})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/webhooks/wh/tok", strings.NewReader(body)))
	return rec
}

func TestWebhookHandlerParsesSlackPayload(t *testing.T) {
	poster := &stubPoster{}
	rec := postWebhook(t, poster, `{"text":"hello","username":"CI","icon_url":"https://example.com/a.png"}`)
	if rec.Code != http.StatusOK || rec.Body.String() != "ok" {
		t.Fatalf("200 ok を期待しましたが %d %s でした", rec.Code, rec.Body.String())
	}
	got := poster.got
	if got.WebhookID != "wh" || got.Token != "tok" || got.Text != "hello" || *got.Username != "CI" || *got.AvatarURL != "https://example.com/a.png" {
		t.Fatalf("入力が正しく渡されていません: %+v", got)
	}
}

func TestWebhookHandlerStatuses(t *testing.T) {
	cases := []struct {
		body string
		err  error
		want int
	}{
		{`not json`, nil, http.StatusBadRequest},
		{`{"text":"a"}`, webhookuc.ErrWebhookNotFound, http.StatusNotFound},
		{`{"text":""}`, webhookuc.ErrEmptyText, http.StatusBadRequest},
		{`{"text":"a"}`, webhookuc.ErrInactive, http.StatusForbidden},
		{`{"text":"a"}`, domerr.ErrChannelArchived, http.StatusForbidden},
		{`{"text":"` + strings.Repeat("a", maxWebhookPayloadBytes) + `"}`, nil, http.StatusRequestEntityTooLarge},
	}
	for _, tc := range cases {
		if rec := postWebhook(t, &stubPoster{err: tc.err}, tc.body); rec.Code != tc.want {
			t.Errorf("%v: %d を期待しましたが %d でした", tc.err, tc.want, rec.Code)
		}
	}
}

func TestRateLimiter(t *testing.T) {
	now := time.Unix(0, 0)
	limiter := newRateLimiter(1, 2)
	limiter.now = func() time.Time { return now }

	for i := range 2 {
		if ok, _ := limiter.allow("a"); !ok {
			t.Fatalf("%d 回目が拒否されました", i+1)
		}
	}
	ok, wait := limiter.allow("a")
	if ok || wait != time.Second {
		t.Fatalf("上限を超えたら 1 秒待たせるはず: ok=%v wait=%v", ok, wait)
	}
	if ok, _ := limiter.allow("b"); !ok {
		t.Fatal("別の Webhook は制限されないはず")
	}
	now = now.Add(time.Second)
	if ok, _ := limiter.allow("a"); !ok {
		t.Fatal("時間が経てば再び受け付けるはず")
	}
}
