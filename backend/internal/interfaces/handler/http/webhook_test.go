package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"

	domerr "github.com/newt239/chat/internal/domain/errors"
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

func postWebhook(t *testing.T, poster *stubPoster, body string) *httptest.ResponseRecorder {
	t.Helper()
	e := NewRouter(RouterConfig{WebhookPoster: poster})
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
		{`{"text":"a"}`, appuc.ErrInactive, http.StatusForbidden},
		{`{"text":"a"}`, appuc.ErrForbiddenChannel, http.StatusForbidden},
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
		if ok, _ := limiter.Allow(t.Context(), "a"); !ok {
			t.Fatalf("%d 回目が拒否されました", i+1)
		}
	}
	ok, wait := limiter.Allow(t.Context(), "a")
	if ok || wait != time.Second {
		t.Fatalf("上限を超えたら 1 秒待たせるはず: ok=%v wait=%v", ok, wait)
	}
	if ok, _ := limiter.Allow(t.Context(), "b"); !ok {
		t.Fatal("別の Webhook は制限されないはず")
	}
	now = now.Add(time.Second)
	if ok, _ := limiter.Allow(t.Context(), "a"); !ok {
		t.Fatal("時間が経てば再び受け付けるはず")
	}
}

func TestWebhookRateLimitCountsOnlyAuthenticatedRequests(t *testing.T) {
	e := NewRouter(RouterConfig{WebhookPoster: &stubPoster{}, WebhookRateLimiter: newRateLimiter(0.001, 1)})
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
