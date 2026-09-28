package entity

import "time"

// UserNote は OwnerID のユーザーだけに見える TargetID のユーザーのニックネームとメモです
type UserNote struct {
	OwnerID   string
	TargetID  string
	Nickname  *string
	Memo      *string
	UpdatedAt time.Time
}
