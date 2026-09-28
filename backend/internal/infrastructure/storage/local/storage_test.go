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
	storage := New(&Config{
		Dir:             t.TempDir(),
		Secret:          "secret",
		MaxFileSize:     1024,
		UploadExpires:   time.Minute,
		DownloadExpires: time.Minute,
	})
	server := httptest.NewServer(storage)
	t.Cleanup(server.Close)
	storage.config.BaseURL = server.URL
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

	uploadURL, err := storage.GenerateUploadURL(key, "image/png", 5, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res := do(t, http.MethodPut, uploadURL, "image/png", "hello"); res.StatusCode != http.StatusOK {
		t.Fatalf("upload status = %d", res.StatusCode)
	}

	downloadURL, err := storage.GenerateDownloadURL(key, nil)
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

	if err := storage.DeleteObject(key); err != nil {
		t.Fatal(err)
	}
	if res := do(t, http.MethodGet, downloadURL, "", ""); res.StatusCode != http.StatusNotFound {
		t.Fatalf("after delete = %d", res.StatusCode)
	}
}

func TestRejectsInvalidRequests(t *testing.T) {
	storage, server := newTestStorage(t)

	downloadURL, _ := storage.GenerateDownloadURL("a/b", nil)
	if res := do(t, http.MethodGet, strings.Replace(downloadURL, "sig=", "sig=0", 1), "", ""); res.StatusCode != http.StatusForbidden {
		t.Fatalf("tampered signature = %d", res.StatusCode)
	}

	storage.now = func() time.Time { return time.Now().Add(2 * time.Minute) }
	if res := do(t, http.MethodGet, downloadURL, "", ""); res.StatusCode != http.StatusForbidden {
		t.Fatalf("expired = %d", res.StatusCode)
	}

	if _, err := storage.GenerateUploadURL("../escape", "", 0, nil); err == nil {
		t.Fatal("path traversal key must be rejected")
	}
	if res := do(t, http.MethodGet, server.URL+"/storage/a/../../x?op=get", "", ""); res.StatusCode != http.StatusForbidden {
		t.Fatalf("unsigned traversal = %d", res.StatusCode)
	}

	uploadURL, _ := storage.GenerateUploadURL("big", "", 0, nil)
	if res := do(t, http.MethodPut, uploadURL, "", strings.Repeat("x", 2048)); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("too large = %d", res.StatusCode)
	}
}
