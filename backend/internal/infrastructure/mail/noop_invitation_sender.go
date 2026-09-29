package mail

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

// NoopInvitationSender はメールを送らない実装です。招待リンクは管理画面でコピーして共有します
type NoopInvitationSender struct{}

func (NoopInvitationSender) SendInvitation(context.Context, *entity.Invitation, string) error {
	return nil
}
