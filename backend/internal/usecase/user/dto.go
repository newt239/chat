package user

type UpdateMeInput struct {
	UserID      string
	DisplayName *string
	Bio         *string
	AvatarURL   *string
}

// MeOutput は自分のプロフィールの出力です
type MeOutput struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	DisplayName string  `json:"displayName"`
	Bio         *string `json:"bio"`
	AvatarURL   *string `json:"avatarUrl"`
}
