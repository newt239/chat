package presenter

import (
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	authuc "github.com/newt239/chat/internal/usecase/auth"
	useruc "github.com/newt239/chat/internal/usecase/user"
)

func AuthUser(u authuc.UserInfo) *chatv1.User {
	return &chatv1.User{Id: u.ID, Email: u.Email, DisplayName: u.DisplayName, AvatarUrl: u.AvatarURL}
}

func Me(me *useruc.MeOutput) *chatv1.User {
	return &chatv1.User{Id: me.ID, Email: me.Email, DisplayName: me.DisplayName, AvatarUrl: me.AvatarURL, Bio: me.Bio}
}
