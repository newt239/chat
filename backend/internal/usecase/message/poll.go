package message

import (
	"strings"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
)

var (
	ErrPollQuestionRequired = domerr.New(domerr.ErrValidation, "投票の質問を入力してください")
	ErrPollOptionInvalid    = domerr.New(domerr.ErrValidation, "選択肢の内容が正しくありません")
	ErrPollClosesInPast     = domerr.New(domerr.ErrValidation, "締切は今より後にしてください")
)

type PollInput struct {
	Question      string
	Mode          entity.PollMode
	AllowMultiple bool
	Anonymous     bool
	ClosesAt      *time.Time
	Options       []PollOptionInput
}

// PollOptionInput は文字の選択肢なら Label、日程調整なら StartsAt を指定します
type PollOptionInput struct {
	Label    string
	StartsAt *time.Time
	AllDay   bool
}

type PollOutput struct {
	ID            string
	Question      string
	Mode          entity.PollMode
	AllowMultiple bool
	Anonymous     bool
	ClosesAt      *time.Time
	IsClosed      bool
	Options       []PollOptionOutput
	// 閲覧者が選んでいる選択肢。配信するときは閲覧者ごとに違うため外す
	MyOptionIDs []string
	VoterCount  int
}

type PollOptionOutput struct {
	ID        string
	Label     string
	StartsAt  *time.Time
	AllDay    bool
	VoteCount int
	// 匿名の投票では返さない
	VoterIDs []string
}

// newPoll は入力を検証して投票を作ります
func newPoll(input *PollInput, now time.Time) (*entity.Poll, error) {
	question := strings.TrimSpace(input.Question)
	if question == "" {
		return nil, ErrPollQuestionRequired
	}
	if input.ClosesAt != nil && !input.ClosesAt.After(now) {
		return nil, ErrPollClosesInPast
	}
	poll := &entity.Poll{
		Question:      question,
		Mode:          input.Mode,
		AllowMultiple: input.AllowMultiple,
		Anonymous:     input.Anonymous,
		ClosesAt:      input.ClosesAt,
	}
	for _, o := range input.Options {
		label := strings.TrimSpace(o.Label)
		if input.Mode == entity.PollModeDate && o.StartsAt == nil ||
			input.Mode == entity.PollModeText && (label == "" || o.StartsAt != nil) {
			return nil, ErrPollOptionInvalid
		}
		poll.Options = append(poll.Options, entity.PollOption{Label: label, StartsAt: o.StartsAt, AllDay: o.AllDay})
	}
	return poll, nil
}

func buildPollOutput(poll *entity.Poll, votes []*entity.PollVote, viewerID string, now time.Time) *PollOutput {
	output := &PollOutput{
		ID:            poll.ID,
		Question:      poll.Question,
		Mode:          poll.Mode,
		AllowMultiple: poll.AllowMultiple,
		Anonymous:     poll.Anonymous,
		ClosesAt:      poll.ClosesAt,
		IsClosed:      poll.IsClosed(now),
		MyOptionIDs:   []string{},
	}
	voters := map[string]bool{}
	for _, o := range poll.Options {
		option := PollOptionOutput{ID: o.ID, Label: o.Label, StartsAt: o.StartsAt, AllDay: o.AllDay, VoterIDs: []string{}}
		for _, v := range votes {
			if v.OptionID != o.ID {
				continue
			}
			option.VoteCount++
			voters[v.UserID] = true
			if !poll.Anonymous {
				option.VoterIDs = append(option.VoterIDs, v.UserID)
			}
			if v.UserID == viewerID {
				output.MyOptionIDs = append(output.MyOptionIDs, o.ID)
			}
		}
		output.Options = append(output.Options, option)
	}
	output.VoterCount = len(voters)
	return output
}
