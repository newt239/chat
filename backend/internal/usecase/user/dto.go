package user

type UpdateMeInput struct {
    UserID      string
    DisplayName *string
    Bio         *string
    AvatarURL   *string
}

type UpdateMeOutput struct {
    ID          string
    DisplayName string
    Bio         *string
    AvatarURL   *string
}


