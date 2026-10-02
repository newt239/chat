package presenter

import (
	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

var pollModes = map[entity.PollMode]chatv1.PollMode{
	entity.PollModeText: chatv1.PollMode_POLL_MODE_TEXT,
	entity.PollModeDate: chatv1.PollMode_POLL_MODE_DATE,
}

func Poll(p *messageuc.PollOutput) *chatv1.Poll {
	if p == nil {
		return nil
	}
	return &chatv1.Poll{
		Id:            p.ID,
		Question:      p.Question,
		Mode:          pollModes[p.Mode],
		AllowMultiple: p.AllowMultiple,
		Anonymous:     p.Anonymous,
		ClosesAt:      optionalTimestamp(p.ClosesAt),
		IsClosed:      p.IsClosed,
		Options: ConvertAll(p.Options, func(o messageuc.PollOptionOutput) *chatv1.PollOption {
			return &chatv1.PollOption{
				Id:        o.ID,
				Label:     o.Label,
				StartsAt:  optionalTimestamp(o.StartsAt),
				AllDay:    o.AllDay,
				VoteCount: int32(o.VoteCount),
				VoterIds:  o.VoterIDs,
			}
		}),
		MyOptionIds: p.MyOptionIDs,
		VoterCount:  int32(p.VoterCount),
	}
}

// PollInput はリクエストの投票を読み替えます。値の範囲は protovalidate で検証済み
func PollInput(p *chatv1.PollInput) *messageuc.PollInput {
	if p == nil {
		return nil
	}
	input := &messageuc.PollInput{
		Question:      p.Question,
		AllowMultiple: p.AllowMultiple,
		Anonymous:     p.Anonymous,
	}
	for mode, v := range pollModes {
		if v == p.Mode {
			input.Mode = mode
		}
	}
	if p.ClosesAt != nil {
		closesAt := p.ClosesAt.AsTime()
		input.ClosesAt = &closesAt
	}
	for _, o := range p.Options {
		option := messageuc.PollOptionInput{Label: o.Label, AllDay: o.AllDay}
		if o.StartsAt != nil {
			startsAt := o.StartsAt.AsTime()
			option.StartsAt = &startsAt
		}
		input.Options = append(input.Options, option)
	}
	return input
}
