package reaction

import "time"

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

type UserInfo struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
}

type ReactionOutput struct {
	MessageID string    `json:"messageId"`
	User      UserInfo  `json:"user"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"createdAt"`
}

type ListReactionsOutput struct {
	Reactions []ReactionOutput `json:"reactions"`
}
