package rpc

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/usecase/audit"
	notificationuc "github.com/newt239/chat/internal/usecase/notification"
)

type NotificationServer struct {
	UC notificationuc.UseCase
}

var pushPlatforms = map[chatv1.PushPlatform]entity.PushPlatform{
	chatv1.PushPlatform_PUSH_PLATFORM_WEB:     entity.PushPlatformWeb,
	chatv1.PushPlatform_PUSH_PLATFORM_IOS:     entity.PushPlatformIOS,
	chatv1.PushPlatform_PUSH_PLATFORM_ANDROID: entity.PushPlatformAndroid,
}

func (s *NotificationServer) RegisterPushToken(ctx context.Context, req *chatv1.RegisterPushTokenRequest) (*chatv1.RegisterPushTokenResponse, error) {
	err := s.UC.RegisterPushToken(ctx, notificationuc.RegisterPushTokenInput{
		UserID:    userIDFrom(ctx),
		Token:     req.Token,
		Platform:  pushPlatforms[req.Platform],
		UserAgent: audit.ClientInfoFrom(ctx).UserAgent,
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.RegisterPushTokenResponse{}, nil
}

func (s *NotificationServer) UnregisterPushToken(ctx context.Context, req *chatv1.UnregisterPushTokenRequest) (*chatv1.UnregisterPushTokenResponse, error) {
	if err := s.UC.UnregisterPushToken(ctx, notificationuc.UnregisterPushTokenInput{UserID: userIDFrom(ctx), Token: req.Token}); err != nil {
		return nil, err
	}
	return &chatv1.UnregisterPushTokenResponse{}, nil
}
