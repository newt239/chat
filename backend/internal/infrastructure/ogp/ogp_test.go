package ogp

import (
	"context"
	"errors"
	"net"
	"testing"

	domainerrors "github.com/newt239/chat/internal/domain/errors"
)

func TestIsBlockedIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{ip: "127.0.0.1", want: true},
		{ip: "10.0.0.1", want: true},
		{ip: "192.168.1.1", want: true},
		{ip: "172.16.0.1", want: true},
		{ip: "169.254.169.254", want: true},
		{ip: "::1", want: true},
		{ip: "0.0.0.0", want: true},
		{ip: "93.184.216.34", want: false},
		{ip: "2606:2800:220:1:248:1893:25c8:1946", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			if got := isBlockedIP(net.ParseIP(tt.ip)); got != tt.want {
				t.Errorf("%s の遮断判定が期待と異なります: got=%t want=%t", tt.ip, got, tt.want)
			}
		})
	}
}

func TestFetchOGPBlocksInternalAddress(t *testing.T) {
	_, err := NewOGPService().FetchOGP(context.Background(), "http://127.0.0.1/")

	if !errors.Is(err, domainerrors.ErrValidation) {
		t.Fatalf("内部ネットワーク宛が遮断されていません: %v", err)
	}
}

func TestFetchOGPRejectsUnsupportedScheme(t *testing.T) {
	_, err := NewOGPService().FetchOGP(context.Background(), "file:///etc/passwd")

	if err == nil {
		t.Fatal("http/https 以外のスキームが拒否されていません")
	}
}
