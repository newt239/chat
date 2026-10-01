package entity

import (
	"regexp"
	"slices"
	"time"
)

type MessageUserMention struct {
	MessageID string
	UserID    string
	// グループへのメンションを投稿時点のメンバーに展開したときの展開元
	ViaGroupID *string
	CreatedAt  time.Time
}

type MessageGroupMention struct {
	MessageID string
	GroupID   string
	CreatedAt time.Time
}

// MentionKind は本文に埋め込む ID 記法の種類です
type MentionKind int

const (
	// MentionKindUser は <@ユーザーID>
	MentionKindUser MentionKind = iota
	// MentionKindGroup は <@&グループID>
	MentionKindGroup
	// MentionKindChannel は <#チャンネルID>
	MentionKindChannel
	// MentionKindBroadcast は <@channel> と <@here>。ID には channel か here が入る
	MentionKindBroadcast
)

const (
	BroadcastChannel = "channel"
	BroadcastHere    = "here"
)

// 名前は変わるため、本文には ID を埋め込む
var mentionTokenPattern = regexp.MustCompile(`<(@&|@|#)([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}|channel|here)>`)

func mentionKindOf(prefix, id string) (MentionKind, bool) {
	isBroadcast := id == BroadcastChannel || id == BroadcastHere
	switch {
	case prefix == "@" && isBroadcast:
		return MentionKindBroadcast, true
	case isBroadcast:
		return 0, false
	case prefix == "@":
		return MentionKindUser, true
	case prefix == "@&":
		return MentionKindGroup, true
	default:
		return MentionKindChannel, true
	}
}

// MentionTokens は本文に埋め込まれた ID 記法を種類ごとに重複なく集めたものです
type MentionTokens struct {
	UserIDs    []string
	GroupIDs   []string
	ChannelIDs []string
	Channel    bool
	Here       bool
}

func ParseMentionTokens(body string) MentionTokens {
	var tokens MentionTokens
	ReplaceMentionTokens(body, func(kind MentionKind, id string) string {
		switch kind {
		case MentionKindUser:
			tokens.UserIDs = appendUnique(tokens.UserIDs, id)
		case MentionKindGroup:
			tokens.GroupIDs = appendUnique(tokens.GroupIDs, id)
		case MentionKindChannel:
			tokens.ChannelIDs = appendUnique(tokens.ChannelIDs, id)
		case MentionKindBroadcast:
			tokens.Channel = tokens.Channel || id == BroadcastChannel
			tokens.Here = tokens.Here || id == BroadcastHere
		}
		return ""
	})
	return tokens
}

// ReplaceMentionTokens は本文の ID 記法を replace の戻り値に置き換えます
func ReplaceMentionTokens(body string, replace func(kind MentionKind, id string) string) string {
	return mentionTokenPattern.ReplaceAllStringFunc(body, func(token string) string {
		m := mentionTokenPattern.FindStringSubmatch(token)
		kind, ok := mentionKindOf(m[1], m[2])
		if !ok {
			return token
		}
		return replace(kind, m[2])
	})
}

func appendUnique(ids []string, id string) []string {
	if slices.Contains(ids, id) {
		return ids
	}
	return append(ids, id)
}
