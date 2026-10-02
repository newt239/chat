package entity

import (
	"slices"
	"testing"
)

const (
	userA   = "11111111-1111-4111-8111-111111111111"
	groupA  = "22222222-2222-4222-8222-222222222222"
	channel = "33333333-3333-4333-8333-333333333333"
)

func TestParseMentionTokens(t *testing.T) {
	got := ParseMentionTokens("<@" + userA + "> <@&" + groupA + "> <#" + channel + "> <@" + userA + "> <@here> <#here> @name")

	if !slices.Equal(got.UserIDs, []string{userA}) || !slices.Equal(got.GroupIDs, []string{groupA}) || !slices.Equal(got.ChannelIDs, []string{channel}) {
		t.Fatalf("ID の抽出が正しくありません: %+v", got)
	}
	if got.Channel || !got.Here {
		t.Fatalf("@channel / @here の判定が正しくありません: %+v", got)
	}
}

func TestReplaceMentionTokens(t *testing.T) {
	got := ReplaceMentionTokens("hi <@"+userA+"> in <#"+channel+"> <#here>", func(kind MentionKind, id string) string {
		if kind == MentionKindUser {
			return "@alice"
		}
		return "#general"
	})

	if want := "hi @alice in #general <#here>"; got != want {
		t.Fatalf("got=%q want=%q", got, want)
	}
}
