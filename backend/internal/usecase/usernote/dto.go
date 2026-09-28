package usernote

import "time"

type GetInput struct {
	OwnerID  string
	TargetID string
}

type UpdateInput struct {
	OwnerID  string
	TargetID string
	// 空文字は未設定として扱う
	Nickname string
	Memo     string
}

type Output struct {
	TargetID  string
	Nickname  *string
	Memo      *string
	UpdatedAt time.Time
}
