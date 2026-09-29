package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type stubOAuthFlow struct{ disabled bool }

func (f stubOAuthFlow) AuthCodeURL(state, codeChallenge, nonce string) (string, error) {
	if f.disabled {
		return "", errors.New("disabled")
	}
	return "https://accounts.google.com/auth?state=" + state, nil
}

func (f stubOAuthFlow) AppRedirectURL(query url.Values) (string, error) {
	return "app://callback?code=" + query.Get("code"), nil
}

func getOAuth(t *testing.T, flow GoogleOAuthFlow, target string) *httptest.ResponseRecorder {
	t.Helper()
	e := NewRouter(RouterConfig{GoogleOAuth: flow})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func TestGoogleOAuthStartRedirectsToGoogle(t *testing.T) {
	q := url.Values{"state": {strings.Repeat("s", 16)}, "code_challenge": {strings.Repeat("c", 43)}, "nonce": {strings.Repeat("n", 16)}}
	rec := getOAuth(t, stubOAuthFlow{}, "/oauth/google/start?"+q.Encode())
	if rec.Code != http.StatusFound || !strings.HasPrefix(rec.Header().Get("Location"), "https://accounts.google.com/") {
		t.Fatalf("Google へのリダイレクトを期待しましたが %d %s でした", rec.Code, rec.Header().Get("Location"))
	}
}

func TestGoogleOAuthStartRejectsShortChallenge(t *testing.T) {
	q := url.Values{"state": {strings.Repeat("s", 16)}, "code_challenge": {"short"}, "nonce": {strings.Repeat("n", 16)}}
	if rec := getOAuth(t, stubOAuthFlow{}, "/oauth/google/start?"+q.Encode()); rec.Code != http.StatusBadRequest {
		t.Fatalf("400 を期待しましたが %d でした", rec.Code)
	}
}

func TestGoogleOAuthCallbackRedirectsToApp(t *testing.T) {
	rec := getOAuth(t, stubOAuthFlow{}, "/oauth/google/callback?code=abc&state=s")
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "app://callback?code=abc" {
		t.Fatalf("アプリへのリダイレクトを期待しましたが %d %s でした", rec.Code, rec.Header().Get("Location"))
	}
}

func TestGoogleOAuthRoutesAbsentWithoutFlow(t *testing.T) {
	if rec := getOAuth(t, nil, "/oauth/google/callback?code=abc"); rec.Code != http.StatusNotFound {
		t.Fatalf("404 を期待しましたが %d でした", rec.Code)
	}
}
