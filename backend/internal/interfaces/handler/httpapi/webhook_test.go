package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/infrastructure/redis"
	appuc "github.com/newt239/chat/internal/usecase/app"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

type stubPoster struct {
	got     appuc.PostInput
	authErr error
	err     error
}

func (p *stubPoster) Authenticate(_ context.Context, appID, token string) (*entity.App, error) {
	if token != "tok" {
		return nil, appuc.ErrAppNotFound
	}
	return &entity.App{ID: appID}, p.authErr
}

func (p *stubPoster) Post(_ context.Context, _ *entity.App, input appuc.PostInput) (*messageuc.MessageOutput, error) {
	p.got = input
	return &messageuc.MessageOutput{}, p.err
}

func newLimiter(t *testing.T, perSecond float64, burst int) *redis.RateLimiter {
	t.Helper()
	client := goredis.NewClient(&goredis.Options{Addr: miniredis.RunT(t).Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return redis.NewRateLimiter(client, "webhook", perSecond, burst)
}

func postWebhook(t *testing.T, poster *stubPoster, body string) *httptest.ResponseRecorder {
	t.Helper()
	e := NewRouter(RouterConfig{WebhookPoster: poster, WebhookRateLimiter: newLimiter(t, WebhookRatePerSecond, WebhookBurst)})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/webhooks/wh/tok", strings.NewReader(body)))
	return rec
}

func TestWebhookHandlerParsesSlackPayload(t *testing.T) {
	poster := &stubPoster{}
	rec := postWebhook(t, poster, `{"text":"hello","channel_id":"c1","thread_id":"m1"}`)
	if rec.Code != http.StatusOK || rec.Body.String() != "ok" {
		t.Fatalf("200 ok を期待しましたが %d %s でした", rec.Code, rec.Body.String())
	}
	got := poster.got
	if got.Text != "hello" ||
		*got.ChannelID != "c1" || *got.ParentID != "m1" {
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
		{`{"text":"a"}`, appuc.ErrAppNotFound, http.StatusNotFound},
		{`{"text":"a"}`, appuc.ErrInactive, http.StatusForbidden},
		{`{"text":""}`, appuc.ErrEmptyText, http.StatusBadRequest},
		{`{"text":"a"}`, appuc.ErrForbiddenChannel, http.StatusForbidden},
		{`{"text":"` + strings.Repeat("a", maxWebhookPayloadBytes) + `"}`, nil, http.StatusRequestEntityTooLarge},
	}
	for _, tc := range cases {
		if rec := postWebhook(t, &stubPoster{err: tc.err}, tc.body); rec.Code != tc.want {
			t.Errorf("%v: %d を期待しましたが %d でした", tc.err, tc.want, rec.Code)
		}
	}
}

func TestWebhookRateLimitCountsOnlyAuthenticatedRequests(t *testing.T) {
	e := NewRouter(RouterConfig{WebhookPoster: &stubPoster{}, WebhookRateLimiter: newLimiter(t, 0.001, 1)})
	send := func(token string) int {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/webhooks/wh/"+token, strings.NewReader(`{"text":"a"}`)))
		return rec.Code
	}
	for range 3 {
		if code := send("wrong"); code != http.StatusNotFound {
			t.Fatalf("不正なトークンは 404 のはず: %d", code)
		}
	}
	if code := send("tok"); code != http.StatusOK {
		t.Fatalf("不正なトークンで枠を使い切られています: %d", code)
	}
	if code := send("tok"); code != http.StatusTooManyRequests {
		t.Fatalf("上限を超えたら 429 のはず: %d", code)
	}
}
