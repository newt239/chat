// Package local は開発用のストレージ。S3 互換のストレージと同じく署名付き URL で読み書きする
package local

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	opGet = "get"
	opPut = "put"
	// Content-Type はファイルの横に保存し、配信時に返す
	contentTypeSuffix = ".content-type"
)

type Storage struct {
	dir     string
	baseURL string
	secret  string
	now     func() time.Time
}

// New は dir に保存し、baseURL を起点に secret で署名した URL を発行します
func New(dir, baseURL, secret string) *Storage {
	return &Storage{dir: dir, baseURL: baseURL, secret: secret, now: time.Now}
}

func (s *Storage) GenerateUploadURL(_ context.Context, key, _ string, sizeBytes int64, expires time.Duration) (string, error) {
	return s.signedURL(opPut, key, strconv.FormatInt(sizeBytes, 10), expires)
}

func (s *Storage) GenerateDownloadURL(_ context.Context, key string, expires time.Duration) (string, error) {
	return s.signedURL(opGet, key, "", expires)
}

func (s *Storage) DeleteObject(_ context.Context, key string) error {
	path, err := s.path(key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	_ = os.Remove(path + contentTypeSuffix)
	return nil
}

// ServeHTTP は /storage/{key} への署名付きの GET / PUT を処理する
func (s *Storage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/storage/")
	op := map[string]string{http.MethodGet: opGet, http.MethodHead: opGet, http.MethodPut: opPut}[r.Method]
	if op == "" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.verify(op, key, r.URL.Query()) {
		http.Error(w, "invalid signature", http.StatusForbidden)
		return
	}
	path, err := s.path(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if op == opPut {
		size, err := strconv.ParseInt(r.URL.Query().Get("size"), 10, 64)
		if err != nil || r.ContentLength != size {
			http.Error(w, "content length does not match the signed size", http.StatusBadRequest)
			return
		}
		s.put(w, r, path)
		return
	}
	file, err := os.Open(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer func() { _ = file.Close() }()
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if contentType, err := os.ReadFile(path + contentTypeSuffix); err == nil {
		w.Header().Set("Content-Type", string(contentType))
	}
	http.ServeContent(w, r, "", stat.ModTime(), file)
}

func (s *Storage) put(w http.ResponseWriter, r *http.Request, path string) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	file, err := os.Create(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(file, r.Body)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(path)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := os.WriteFile(path+contentTypeSuffix, []byte(r.Header.Get("Content-Type")), 0o644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// signedURL は書き込みなら size も署名に含め、本文の大きさを変えられないようにする
func (s *Storage) signedURL(op, key, size string, expires time.Duration) (string, error) {
	if _, err := s.path(key); err != nil {
		return "", err
	}
	exp := strconv.FormatInt(s.now().Add(expires).Unix(), 10)
	query := url.Values{"exp": {exp}, "op": {op}, "sig": {s.sign(op, key, exp, size)}}
	if size != "" {
		query.Set("size", size)
	}
	return fmt.Sprintf("%s/storage/%s?%s", strings.TrimRight(s.baseURL, "/"), key, query.Encode()), nil
}

func (s *Storage) verify(op, key string, query url.Values) bool {
	exp, err := strconv.ParseInt(query.Get("exp"), 10, 64)
	if err != nil || s.now().Unix() > exp || query.Get("op") != op {
		return false
	}
	return hmac.Equal([]byte(query.Get("sig")), []byte(s.sign(op, key, query.Get("exp"), query.Get("size"))))
}

func (s *Storage) sign(op, key, exp, size string) string {
	mac := hmac.New(sha256.New, []byte(s.secret))
	mac.Write([]byte(op + "\n" + key + "\n" + exp + "\n" + size))
	return hex.EncodeToString(mac.Sum(nil))
}

// path はキーを保存先のパスに変換する。ディレクトリの外を指すキーは拒否する
func (s *Storage) path(key string) (string, error) {
	cleaned := filepath.Clean("/" + key)
	if key == "" || cleaned != "/"+key || strings.HasSuffix(key, contentTypeSuffix) {
		return "", fmt.Errorf("invalid storage key: %q", key)
	}
	return filepath.Join(s.dir, filepath.FromSlash(cleaned)), nil
}
