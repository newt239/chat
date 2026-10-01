package safehttp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// ErrBlockedAddress は内部ネットワーク宛のリクエストを拒否したことを表します
var ErrBlockedAddress = errors.New("内部ネットワーク宛の URL には接続できません")

// isBlockedIP はループバック・プライベート・リンクローカル（クラウドのメタデータ含む）を弾きます
func isBlockedIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast()
}

// NewClient は内部ネットワーク宛の接続を拒否する HTTP クライアントを作ります
func NewClient(timeout time.Duration, maxRedirects int) *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return fmt.Errorf("リダイレクトが多すぎます")
			}
			if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
				return fmt.Errorf("unsupported URL scheme: %s", req.URL.Scheme)
			}
			return nil
		},
		// 名前解決後のアドレスを検証し、DNS リバインディングによる迂回も防ぐ
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
				if err != nil {
					return nil, err
				}
				for _, ip := range ips {
					if isBlockedIP(ip) {
						return nil, ErrBlockedAddress
					}
				}
				return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].String(), port))
			},
		},
	}
}
