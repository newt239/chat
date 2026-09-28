// Package local は開発用のストレージ。署名付き URL で読み書きする点は S3 互換のストレージと同じにし、
// フロントエンドはどちらでも同じ手順でアップロード・表示できる
package local

import (
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

type Config struct {
	Dir             string
	BaseURL         string
	Secret          string
	MaxFileSize     int64
	UploadExpires   time.Duration
	DownloadExpires time.Duration
}

type Storage struct {
	config *Config
	now    func() time.Time
}

func New(cfg *Config) *Storage {
	return &Storage{config: cfg, now: time.Now}
}

func (s *Storage) GetMaxFileSize() int64 { return s.config.MaxFileSize }

func (s *Storage) GetUploadExpires() interface{} { return s.config.UploadExpires }

func (s *Storage) GetDownloadExpires() interface{} { return s.config.DownloadExpires }

func (s *Storage) GenerateUploadURL(key, _ string, _ int64, expires interface{}) (string, error) {
	return s.signedURL(opPut, key, durationOr(expires, s.config.UploadExpires))
}

func (s *Storage) GenerateDownloadURL(key string, expires interface{}) (string, error) {
	return s.signedURL(opGet, key, durationOr(expires, s.config.DownloadExpires))
}

func (s *Storage) DeleteObject(key string) error {
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
	_, err = io.Copy(file, http.MaxBytesReader(w, r.Body, s.config.MaxFileSize))
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

func (s *Storage) signedURL(op, key string, expires time.Duration) (string, error) {
	if _, err := s.path(key); err != nil {
		return "", err
	}
	exp := strconv.FormatInt(s.now().Add(expires).Unix(), 10)
	query := url.Values{"exp": {exp}, "op": {op}, "sig": {s.sign(op, key, exp)}}
	return fmt.Sprintf("%s/storage/%s?%s", strings.TrimRight(s.config.BaseURL, "/"), key, query.Encode()), nil
}

func (s *Storage) verify(op, key string, query url.Values) bool {
	exp, err := strconv.ParseInt(query.Get("exp"), 10, 64)
	if err != nil || s.now().Unix() > exp || query.Get("op") != op {
		return false
	}
	return hmac.Equal([]byte(query.Get("sig")), []byte(s.sign(op, key, query.Get("exp"))))
}

func (s *Storage) sign(op, key, exp string) string {
	mac := hmac.New(sha256.New, []byte(s.config.Secret))
	mac.Write([]byte(op + "\n" + key + "\n" + exp))
	return hex.EncodeToString(mac.Sum(nil))
}

// path はキーを保存先のパスに変換する。ディレクトリの外を指すキーは拒否する
func (s *Storage) path(key string) (string, error) {
	cleaned := filepath.Clean("/" + key)
	if key == "" || cleaned != "/"+key || strings.HasSuffix(key, contentTypeSuffix) {
		return "", fmt.Errorf("invalid storage key: %q", key)
	}
	return filepath.Join(s.config.Dir, filepath.FromSlash(cleaned)), nil
}

func durationOr(value interface{}, fallback time.Duration) time.Duration {
	if d, ok := value.(time.Duration); ok && d > 0 {
		return d
	}
	return fallback
}
