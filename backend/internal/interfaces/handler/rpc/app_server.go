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
	out, err := s.UC.Create(ctx, appuc.CreateInput{WorkspaceID: req.WorkspaceId, UserID: userIDFrom(ctx), Settings: presenter.AppSettingsFromProto(req.Settings)})
	if err != nil {
		return nil, err
	}
	return &chatv1.CreateAppResponse{App: presenter.App(out.App), Token: out.Token}, nil
}

func (s *AppServer) UpdateApp(ctx context.Context, req *chatv1.UpdateAppRequest) (*chatv1.UpdateAppResponse, error) {
	out, err := s.UC.Update(ctx, appuc.UpdateInput{AppID: req.AppId, UserID: userIDFrom(ctx), Settings: presenter.AppSettingsFromProto(req.Settings)})
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
	if err := s.UC.Delete(ctx, appuc.TargetInput{AppID: req.AppId, UserID: userIDFrom(ctx)}); err != nil {
		return nil, err
	}
	return &chatv1.DeleteAppResponse{}, nil
}

func (s *AppServer) AddAppToChannel(ctx context.Context, req *chatv1.AddAppToChannelRequest) (*chatv1.AddAppToChannelResponse, error) {
	if err := s.UC.AddToChannel(ctx, appuc.ChannelInput{AppID: req.AppId, ChannelID: req.ChannelId, UserID: userIDFrom(ctx)}); err != nil {
		return nil, err
	}
	return &chatv1.AddAppToChannelResponse{}, nil
}

func (s *AppServer) RemoveAppFromChannel(ctx context.Context, req *chatv1.RemoveAppFromChannelRequest) (*chatv1.RemoveAppFromChannelResponse, error) {
	if err := s.UC.RemoveFromChannel(ctx, appuc.ChannelInput{AppID: req.AppId, ChannelID: req.ChannelId, UserID: userIDFrom(ctx)}); err != nil {
		return nil, err
	}
	return &chatv1.RemoveAppFromChannelResponse{}, nil
}
