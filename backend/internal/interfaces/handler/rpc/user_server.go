package rpc

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	useruc "github.com/newt239/chat/internal/usecase/user"
	usernoteuc "github.com/newt239/chat/internal/usecase/usernote"
)

type UserServer struct {
	UC     *useruc.Interactor
	NoteUC *usernoteuc.Interactor
}

func (s *UserServer) GetMe(ctx context.Context, _ *chatv1.GetMeRequest) (*chatv1.GetMeResponse, error) {
	out, err := s.UC.GetMe(ctx, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.GetMeResponse{User: presenter.Me(out)}, nil
}

func (s *UserServer) UpdateMe(ctx context.Context, req *chatv1.UpdateMeRequest) (*chatv1.UpdateMeResponse, error) {
	input := useruc.UpdateMeInput{UserID: userIDFrom(ctx), DisplayName: req.DisplayName, Bio: req.Bio, AvatarURL: req.AvatarUrl}
	if req.Links != nil {
		input.Links = &req.Links.Urls
	}
	out, err := s.UC.UpdateMe(ctx, input)
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdateMeResponse{User: presenter.Me(out)}, nil
}

func (s *UserServer) UpdatePreferences(ctx context.Context, req *chatv1.UpdatePreferencesRequest) (*chatv1.UpdatePreferencesResponse, error) {
	p := req.Preferences
	out, err := s.UC.UpdatePreferences(ctx, userIDFrom(ctx), entity.UserPreferences{
		ThemeHue:           int(p.GetTheme().GetHue()),
		ThemeChroma:        p.GetTheme().GetChroma(),
		ThemeSidebar:       fromProto(presenter.SidebarStyles, p.GetTheme().GetSidebar()),
		ColorMode:          fromProto(presenter.ColorModes, p.GetColorMode()),
		Locale:             p.GetLocale(),
		NotificationLevel:  fromProto(presenter.NotificationLevels, p.GetNotificationLevel()),
		Timezone:           p.GetTimezone(),
		TimezoneAutoUpdate: p.GetTimezoneAutoUpdate(),
		ChannelSortOrder:   fromProto(presenter.ChannelSortOrders, p.GetChannelSortOrder()),
		HideJoinMessages:   p.GetHideJoinMessages(),
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdatePreferencesResponse{Preferences: presenter.Preferences(*out)}, nil
}

func (s *UserServer) UpdatePassword(ctx context.Context, req *chatv1.UpdatePasswordRequest) (*chatv1.UpdatePasswordResponse, error) {
	return &chatv1.UpdatePasswordResponse{}, s.UC.UpdatePassword(ctx, userIDFrom(ctx), req.CurrentPassword, req.NewPassword)
}

func (s *UserServer) DeleteMe(ctx context.Context, _ *chatv1.DeleteMeRequest) (*chatv1.DeleteMeResponse, error) {
	return &chatv1.DeleteMeResponse{}, s.UC.DeleteMe(ctx, userIDFrom(ctx))
}

func (s *UserServer) GetUserNote(ctx context.Context, req *chatv1.GetUserNoteRequest) (*chatv1.GetUserNoteResponse, error) {
	out, err := s.NoteUC.Get(ctx, userIDFrom(ctx), req.TargetUserId)
	if err != nil {
		return nil, err
	}
	return &chatv1.GetUserNoteResponse{Note: presenter.UserNote(out)}, nil
}

func (s *UserServer) UpdateUserNote(ctx context.Context, req *chatv1.UpdateUserNoteRequest) (*chatv1.UpdateUserNoteResponse, error) {
	out, err := s.NoteUC.Update(ctx, entity.UserNote{OwnerID: userIDFrom(ctx), TargetID: req.TargetUserId, Nickname: &req.Nickname, Memo: &req.Memo})
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdateUserNoteResponse{Note: presenter.UserNote(out)}, nil
}
