package reaction

import (
	"time"

	"github.com/newt239/chat/internal/usecase/message"
)

type AddReactionInput struct {
	MessageID string
	UserID    string
	Emoji     string
}

type RemoveReactionInput struct {
	MessageID string
	UserID    string
	Emoji     string
}

type ReactionOutput struct {
	MessageID string           `json:"messageId"`
	User      message.UserInfo `json:"user"`
	Emoji     string           `json:"emoji"`
	CreatedAt time.Time        `json:"createdAt"`
}

type ListReactionsOutput struct {
	Reactions []ReactionOutput `json:"reactions"`
}
