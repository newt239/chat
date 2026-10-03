package local

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestStorage(t *testing.T) (*Storage, *httptest.Server) {
	t.Helper()
	storage := New(t.TempDir(), "", "secret")
	server := httptest.NewServer(storage)
	t.Cleanup(server.Close)
	storage.baseURL = server.URL
	return storage, server
}

func do(t *testing.T, method, url, contentType, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { res.Body.Close() })
	return res
}

func TestUploadAndDownload(t *testing.T) {
	storage, _ := newTestStorage(t)
	key := "attachments/ch/file"

	uploadURL, err := storage.GenerateUploadURL(t.Context(), key, "image/png", 5, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if res := do(t, http.MethodPut, uploadURL, "image/png", "hello"); res.StatusCode != http.StatusOK {
		t.Fatalf("upload status = %d", res.StatusCode)
	}

	downloadURL, err := storage.GenerateDownloadURL(t.Context(), key, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	res := do(t, http.MethodGet, downloadURL, "", "")
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK || string(body) != "hello" {
		t.Fatalf("download = %d %q", res.StatusCode, body)
	}
	if got := res.Header.Get("Content-Type"); got != "image/png" {
		t.Fatalf("content type = %q", got)
	}

	// 読み出し用の署名では書き込めない
	if res := do(t, http.MethodPut, downloadURL, "", "x"); res.StatusCode != http.StatusForbidden {
		t.Fatalf("put with get signature = %d", res.StatusCode)
	}

	if err := storage.DeleteObject(t.Context(), key); err != nil {
		t.Fatal(err)
	}
	if res := do(t, http.MethodGet, downloadURL, "", ""); res.StatusCode != http.StatusNotFound {
		t.Fatalf("after delete = %d", res.StatusCode)
	}
}

func TestRejectsInvalidRequests(t *testing.T) {
	storage, server := newTestStorage(t)

	downloadURL, _ := storage.GenerateDownloadURL(t.Context(), "a/b", time.Minute)
	if res := do(t, http.MethodGet, strings.Replace(downloadURL, "sig=", "sig=0", 1), "", ""); res.StatusCode != http.StatusForbidden {
		t.Fatalf("tampered signature = %d", res.StatusCode)
	}

	storage.now = func() time.Time { return time.Now().Add(2 * time.Minute) }
	if res := do(t, http.MethodGet, downloadURL, "", ""); res.StatusCode != http.StatusForbidden {
		t.Fatalf("expired = %d", res.StatusCode)
	}

	if _, err := storage.GenerateUploadURL(t.Context(), "../escape", "", 1, time.Minute); err == nil {
		t.Fatal("path traversal key must be rejected")
	}
	if res := do(t, http.MethodGet, server.URL+"/storage/a/../../x?op=get", "", ""); res.StatusCode != http.StatusForbidden {
		t.Fatalf("unsigned traversal = %d", res.StatusCode)
	}
}

func TestRejectsBodyOtherThanSignedSize(t *testing.T) {
	storage, _ := newTestStorage(t)

	uploadURL, err := storage.GenerateUploadURL(t.Context(), "a/b", "image/png", 5, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{"hello world", "hi"} {
		if res := do(t, http.MethodPut, uploadURL, "image/png", body); res.StatusCode != http.StatusBadRequest {
			t.Errorf("署名と違う大きさ %d バイトの本文を受け付けました: %d", len(body), res.StatusCode)
		}
	}
	if res := do(t, http.MethodPut, strings.Replace(uploadURL, "size=5", "size=11", 1), "image/png", "hello world"); res.StatusCode != http.StatusForbidden {
		t.Errorf("署名後に書き換えたサイズを受け付けました: %d", res.StatusCode)
	}
}
