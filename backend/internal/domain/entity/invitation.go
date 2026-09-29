package entity

import (
	"strings"
	"time"
)

const InvitationTTL = 7 * 24 * time.Hour

// Invitation はメールアドレス宛てのワークスペースへの招待です
type Invitation struct {
	ID          string
	WorkspaceID string
	Email       string
	Role        WorkspaceRole
	TokenHash   string
	InvitedBy   string
	ExpiresAt   time.Time
	AcceptedAt  *time.Time
	CreatedAt   time.Time
}

func (i *Invitation) IsPending(now time.Time) bool {
	return i.AcceptedAt == nil && now.Before(i.ExpiresAt)
}

// NormalizeEmail は招待と Google アカウントのメールアドレスを大文字小文字を区別せずに照合するためのものです
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
