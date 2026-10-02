package ogp

import (
	"context"
	"errors"
	"testing"

	domainerrors "github.com/newt239/chat/internal/domain/errors"
)

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
