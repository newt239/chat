package user

type UpdateMeInput struct {
	UserID      string
	DisplayName *string
	Bio         *string
	AvatarURL   *string
}

type UpdateMeOutput struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"displayName"`
	Bio         *string `json:"bio"`
	AvatarURL   *string `json:"avatarUrl"`
}
