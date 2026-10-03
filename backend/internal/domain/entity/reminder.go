package entity

import "time"

// Reminder は /remind で設定したリマインダーです。TargetUserID も TargetChannelID もなければ作成者本人に届ける
type Reminder struct {
	ID              string
	WorkspaceID     string
	CreatorID       string
	TargetUserID    *string
	TargetChannelID *string
	Text            string
	RemindAt        time.Time
}
