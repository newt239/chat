package safehttp

import (
	"net"
	"testing"
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
