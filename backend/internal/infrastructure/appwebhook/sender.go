// Package appwebhook はアプリの送信 Webhook を外部へ送ります
package appwebhook

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/newt239/chat/internal/infrastructure/safehttp"
)

type Sender struct {
	client *http.Client
}

// NewSender は内部ネットワーク宛を拒否し、リダイレクトをたどらないクライアントで送ります
func NewSender() *Sender {
	return &Sender{client: safehttp.NewClient(5*time.Second, 0)}
}

func (s *Sender) Send(ctx context.Context, url string, body []byte, headers map[string]string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("User-Agent", "ChatApp-Webhook/1.0")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}
