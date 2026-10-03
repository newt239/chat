package message

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

func TestNewPollValidates(t *testing.T) {
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	day := now.Add(24 * time.Hour)
	tests := []struct {
		name  string
		input PollInput
		want  error
	}{
		{name: "文字の選択肢", input: PollInput{Question: "昼ご飯", Mode: entity.PollModeText, Options: []PollOptionInput{{Label: "そば"}, {Label: "うどん"}}}},
		{name: "日程調整", input: PollInput{Question: "打ち上げ", Mode: entity.PollModeDate, Options: []PollOptionInput{{StartsAt: &day, AllDay: true}, {StartsAt: &day}}}},
		{name: "質問が空", input: PollInput{Question: " ", Mode: entity.PollModeText, Options: []PollOptionInput{{Label: "a"}, {Label: "b"}}}, want: ErrPollQuestionRequired},
		{name: "日程調整に日時がない", input: PollInput{Question: "q", Mode: entity.PollModeDate, Options: []PollOptionInput{{Label: "a"}, {StartsAt: &day}}}, want: ErrPollOptionInvalid},
		{name: "締切が過去", input: PollInput{Question: "q", Mode: entity.PollModeText, ClosesAt: &past, Options: []PollOptionInput{{Label: "a"}, {Label: "b"}}}, want: ErrPollClosesInPast},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newPoll(&tt.input, now)
			if !errors.Is(err, tt.want) {
				t.Fatalf("got=%v want=%v", err, tt.want)
			}
		})
	}
}

func TestBuildPollOutputHidesVotersWhenAnonymous(t *testing.T) {
	now := time.Now()
	poll := &entity.Poll{ID: "p", Options: []entity.PollOption{{ID: "a"}, {ID: "b"}}}
	votes := []*entity.PollVote{{OptionID: "a", UserID: "u1"}, {OptionID: "b", UserID: "u1"}, {OptionID: "a", UserID: "u2"}}

	open := buildPollOutput(poll, votes, "u1", now)
	if open.VoterCount != 2 || open.Options[0].VoteCount != 2 || !slices.Equal(open.Options[0].VoterIDs, []string{"u1", "u2"}) || !slices.Equal(open.MyOptionIDs, []string{"a", "b"}) {
		t.Fatalf("集計が期待と異なります: %+v", open)
	}

	poll.Anonymous = true
	anonymous := buildPollOutput(poll, votes, "u2", now)
	if len(anonymous.Options[0].VoterIDs) != 0 || anonymous.Options[0].VoteCount != 2 || !slices.Equal(anonymous.MyOptionIDs, []string{"a"}) {
		t.Fatalf("匿名の投票で投票者が見えています: %+v", anonymous)
	}
}
