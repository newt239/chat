package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterMountsRPCHandler(t *testing.T) {
	rpc := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	e := NewRouter(RouterConfig{RPCHandler: rpc})

	tests := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodPost, "/chat.v1.AuthService/Login", http.StatusTeapot},
		{http.MethodGet, "/healthz", http.StatusOK},
		{http.MethodPost, "/api/auth/login", http.StatusNotFound},
	}
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
		if rec.Code != tt.want {
			t.Errorf("%s %s: %d を期待しましたが %d でした", tt.method, tt.path, tt.want, rec.Code)
		}
	}
}
