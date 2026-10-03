package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	appuc "github.com/newt239/chat/internal/usecase/app"
)

type AppServer struct {
	UC *appuc.Interactor
}

// appSettings は未定義の権限を除いて読み替えます。権限の値は protovalidate で検証済み
func appSettings(s *chatv1.AppSettings) appuc.SettingsInput {
	input := appuc.SettingsInput{
		Name:             s.Name,
		Description:      s.Description,
		AvatarURL:        s.AvatarUrl,
		DefaultChannelID: s.DefaultChannelId,
		OutgoingURL:      s.OutgoingUrl,
	}
	for _, p := range s.Permissions {
		if permission := fromProto(presenter.AppPermissions, p); permission != "" {
			input.Permissions = append(input.Permissions, permission)
		}
	}
	return input
}

func (s *AppServer) ListApps(ctx context.Context, req *chatv1.ListAppsRequest) (*chatv1.ListAppsResponse, error) {
	out, err := s.UC.List(ctx, appuc.ListInput{WorkspaceID: req.WorkspaceId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.ListAppsResponse{Apps: presenter.ConvertAll(out, presenter.App)}, nil
}

func (s *AppServer) ListChannelApps(ctx context.Context, req *chatv1.ListChannelAppsRequest) (*chatv1.ListChannelAppsResponse, error) {
	out, err := s.UC.ListByChannel(ctx, appuc.ListByChannelInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.ListChannelAppsResponse{Apps: presenter.ConvertAll(out, presenter.App)}, nil
}

func (s *AppServer) CreateApp(ctx context.Context, req *chatv1.CreateAppRequest) (*chatv1.CreateAppResponse, error) {
	out, err := s.UC.Create(ctx, appuc.CreateInput{WorkspaceID: req.WorkspaceId, UserID: userIDFrom(ctx), Settings: appSettings(req.Settings)})
	if err != nil {
		return nil, err
	}
	return &chatv1.CreateAppResponse{App: presenter.App(out.App), Token: out.Token}, nil
}

func (s *AppServer) UpdateApp(ctx context.Context, req *chatv1.UpdateAppRequest) (*chatv1.UpdateAppResponse, error) {
	out, err := s.UC.Update(ctx, appuc.UpdateInput{AppID: req.AppId, UserID: userIDFrom(ctx), Settings: appSettings(req.Settings)})
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdateAppResponse{App: presenter.App(*out)}, nil
}

func (s *AppServer) RegenerateAppToken(ctx context.Context, req *chatv1.RegenerateAppTokenRequest) (*chatv1.RegenerateAppTokenResponse, error) {
	token, err := s.UC.RegenerateToken(ctx, appuc.TargetInput{AppID: req.AppId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.RegenerateAppTokenResponse{Token: token}, nil
}

func (s *AppServer) DeleteApp(ctx context.Context, req *chatv1.DeleteAppRequest) (*chatv1.DeleteAppResponse, error) {
	return &chatv1.DeleteAppResponse{}, s.UC.Delete(ctx, appuc.TargetInput{AppID: req.AppId, UserID: userIDFrom(ctx)})
}

func (s *AppServer) AddAppToChannel(ctx context.Context, req *chatv1.AddAppToChannelRequest) (*chatv1.AddAppToChannelResponse, error) {
	return &chatv1.AddAppToChannelResponse{}, s.UC.AddToChannel(ctx, appuc.ChannelInput{AppID: req.AppId, ChannelID: req.ChannelId, UserID: userIDFrom(ctx)})
}

func (s *AppServer) RemoveAppFromChannel(ctx context.Context, req *chatv1.RemoveAppFromChannelRequest) (*chatv1.RemoveAppFromChannelResponse, error) {
	return &chatv1.RemoveAppFromChannelResponse{}, s.UC.RemoveFromChannel(ctx, appuc.ChannelInput{AppID: req.AppId, ChannelID: req.ChannelId, UserID: userIDFrom(ctx)})
}
