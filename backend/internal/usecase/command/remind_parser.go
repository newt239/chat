package command

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	domerr "github.com/newt239/chat/internal/domain/errors"
)

// RemindTarget はリマインダーを届ける先です。どちらも nil なら実行した本人に届ける
type RemindTarget struct {
	UserID    *string
	ChannelID *string
}

type RemindRequest struct {
	Target RemindTarget
	Text   string
	At     time.Time
}

const maxRemindAhead = 366 * 24 * time.Hour

var (
	ErrRemindUsage    = fmt.Errorf("%w: 使い方: /remind [me|@ユーザー|#チャンネル] 内容 日時（例: /remind me 資料を送る 明日 9:00）", domerr.ErrValidation)
	ErrRemindInPast   = fmt.Errorf("%w: リマインドする日時は今より後にしてください", domerr.ErrValidation)
	ErrRemindTooFar   = fmt.Errorf("%w: リマインドは 1 年以内にしてください", domerr.ErrValidation)
	remindTargetToken = regexp.MustCompile(`^(?:me|<@([0-9a-f-]{36})>|<#([0-9a-f-]{36})>)(?:\s+|$)`)
	clock             = `(\d{1,2}):(\d{2})`
	// 本文の末尾にある日時。前から順に試す
	remindTimePatterns = []struct {
		pattern *regexp.Regexp
		resolve func(m []string, now time.Time) (time.Time, bool)
	}{
		{regexp.MustCompile(`(\d+)\s*(分|時間|日)後に?$`), func(m []string, now time.Time) (time.Time, bool) {
			return addDuration(now, m[1], map[string]time.Duration{"分": time.Minute, "時間": time.Hour, "日": 24 * time.Hour}[m[2]])
		}},
		{regexp.MustCompile(`(?i)in\s+(\d+)\s*(m|mins?|minutes?|h|hours?|d|days?)$`), func(m []string, now time.Time) (time.Time, bool) {
			unit := map[byte]time.Duration{'m': time.Minute, 'h': time.Hour, 'd': 24 * time.Hour}[strings.ToLower(m[2])[0]]
			return addDuration(now, m[1], unit)
		}},
		{regexp.MustCompile(`(?i)(今日|明日|明後日|today|tomorrow)\s*` + clock + `に?$`), func(m []string, now time.Time) (time.Time, bool) {
			days := map[string]int{"今日": 0, "today": 0, "明日": 1, "tomorrow": 1, "明後日": 2}[strings.ToLower(m[1])]
			return atClock(now.AddDate(0, 0, days), m[2], m[3])
		}},
		{regexp.MustCompile(`(\d{4})[-/](\d{1,2})[-/](\d{1,2})\s+` + clock + `に?$`), func(m []string, now time.Time) (time.Time, bool) {
			return atDate(now, atoi(m[1]), m[2], m[3], m[4], m[5])
		}},
		{regexp.MustCompile(`(\d{1,2})/(\d{1,2})\s+` + clock + `に?$`), func(m []string, now time.Time) (time.Time, bool) {
			at, ok := atDate(now, now.Year(), m[1], m[2], m[3], m[4])
			// 年を省いた日付が過ぎていれば来年とみなす
			if ok && !at.After(now) {
				at = at.AddDate(1, 0, 0)
			}
			return at, ok
		}},
		{regexp.MustCompile(clock + `に?$`), func(m []string, now time.Time) (time.Time, bool) {
			at, ok := atClock(now, m[1], m[2])
			// 時刻だけなら次に来るその時刻
			if ok && !at.After(now) {
				at = at.AddDate(0, 0, 1)
			}
			return at, ok
		}},
	}
)

// ParseRemind は「/remind」に続く引数を読みます。日時は now のタイムゾーンで解釈する
func ParseRemind(args string, now time.Time) (*RemindRequest, error) {
	args = strings.TrimSpace(args)
	req := &RemindRequest{}
	if m := remindTargetToken.FindStringSubmatch(args); m != nil {
		switch {
		case m[1] != "":
			req.Target.UserID = &m[1]
		case m[2] != "":
			req.Target.ChannelID = &m[2]
		}
		args = strings.TrimSpace(args[len(m[0]):])
	}

	for _, p := range remindTimePatterns {
		loc := p.pattern.FindStringSubmatchIndex(args)
		if loc == nil || (loc[0] > 0 && args[loc[0]-1] != ' ' && !strings.HasSuffix(args[:loc[0]], "　")) {
			continue
		}
		m := p.pattern.FindStringSubmatch(args)
		at, ok := p.resolve(m, now)
		if !ok {
			return nil, ErrRemindUsage
		}
		req.Text = trimQuotes(strings.TrimSpace(args[:loc[0]]))
		req.At = at
		break
	}
	if req.Text == "" || req.At.IsZero() {
		return nil, ErrRemindUsage
	}
	if !req.At.After(now) {
		return nil, ErrRemindInPast
	}
	if req.At.Sub(now) > maxRemindAhead {
		return nil, ErrRemindTooFar
	}
	return req, nil
}

func trimQuotes(text string) string {
	for _, pair := range [][2]string{{"「", "」"}, {`"`, `"`}, {"“", "”"}} {
		if strings.HasPrefix(text, pair[0]) && strings.HasSuffix(text, pair[1]) && len(text) > len(pair[0])+len(pair[1]) {
			return strings.TrimSpace(text[len(pair[0]) : len(text)-len(pair[1])])
		}
	}
	return text
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func addDuration(now time.Time, amount string, unit time.Duration) (time.Time, bool) {
	n := atoi(amount)
	return now.Add(time.Duration(n) * unit), n > 0
}

func atClock(day time.Time, hour, minute string) (time.Time, bool) {
	h, m := atoi(hour), atoi(minute)
	if h > 23 || m > 59 {
		return time.Time{}, false
	}
	return time.Date(day.Year(), day.Month(), day.Day(), h, m, 0, 0, day.Location()), true
}

func atDate(now time.Time, year int, month, day, hour, minute string) (time.Time, bool) {
	mo, d := atoi(month), atoi(day)
	at, ok := atClock(time.Date(year, time.Month(mo), d, 0, 0, 0, 0, now.Location()), hour, minute)
	// 2/30 のような存在しない日付は繰り上がるため弾く
	return at, ok && at.Month() == time.Month(mo) && at.Day() == d
}
