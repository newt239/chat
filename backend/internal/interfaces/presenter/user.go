package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	authuc "github.com/newt239/chat/internal/usecase/auth"
	useruc "github.com/newt239/chat/internal/usecase/user"
	usernoteuc "github.com/newt239/chat/internal/usecase/usernote"
)

func AuthUser(u authuc.UserInfo) *chatv1.User {
	return &chatv1.User{Id: u.ID, Email: u.Email, DisplayName: u.DisplayName, AvatarUrl: u.AvatarURL}
}

func Me(me *useruc.MeOutput) *chatv1.User {
	return &chatv1.User{Id: me.ID, Email: me.Email, DisplayName: me.DisplayName, AvatarUrl: me.AvatarURL, Bio: me.Bio}
}

// UserNote は未設定 (nil) の場合 nil を返します
func UserNote(n *usernoteuc.Output) *chatv1.UserNote {
	if n == nil {
		return nil
	}
	return &chatv1.UserNote{TargetUserId: n.TargetID, Nickname: n.Nickname, Memo: n.Memo, UpdatedAt: timestamppb.New(n.UpdatedAt)}
}
