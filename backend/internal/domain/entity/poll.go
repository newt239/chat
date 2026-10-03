package entity

import (
	"slices"
	"time"
)

type PollMode string

const (
	PollModeText PollMode = "text"
	// PollModeDate は日時の候補から選ぶ日程調整
	PollModeDate PollMode = "date"
)

type Poll struct {
	ID            string
	MessageID     string
	Question      string
	Mode          PollMode
	AllowMultiple bool
	// 誰がどれに投票したかを作成者にも見せない
	Anonymous bool
	ClosesAt  *time.Time
	ClosedAt  *time.Time
	// 並び順
	Options []PollOption
}

type PollOption struct {
	ID       string
	Label    string
	StartsAt *time.Time
	AllDay   bool
}

type PollVote struct {
	PollID   string
	OptionID string
	UserID   string
}

// IsClosed は締め切ったか、締切の日時を過ぎたかを返します
func (p *Poll) IsClosed(now time.Time) bool {
	return p.ClosedAt != nil || (p.ClosesAt != nil && !now.Before(*p.ClosesAt))
}

// HasOption は選択肢がこの投票のものかを返します
func (p *Poll) HasOption(optionID string) bool {
	return slices.ContainsFunc(p.Options, func(o PollOption) bool { return o.ID == optionID })
}
