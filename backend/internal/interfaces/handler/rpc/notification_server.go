package rpc

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/usecase/audit"
	notificationuc "github.com/newt239/chat/internal/usecase/notification"
)

type NotificationServer struct {
	UC *notificationuc.Interactor
}

var pushPlatforms = map[chatv1.PushPlatform]entity.PushPlatform{
	chatv1.PushPlatform_PUSH_PLATFORM_WEB:     entity.PushPlatformWeb,
	chatv1.PushPlatform_PUSH_PLATFORM_IOS:     entity.PushPlatformIOS,
	chatv1.PushPlatform_PUSH_PLATFORM_ANDROID: entity.PushPlatformAndroid,
}

func (s *NotificationServer) RegisterPushToken(ctx context.Context, req *chatv1.RegisterPushTokenRequest) (*chatv1.RegisterPushTokenResponse, error) {
	return &chatv1.RegisterPushTokenResponse{}, s.UC.RegisterPushToken(ctx, entity.PushToken{
		UserID:    userIDFrom(ctx),
		Token:     req.Token,
		Platform:  pushPlatforms[req.Platform],
		UserAgent: audit.ClientInfoFrom(ctx).UserAgent,
	})
}

func (s *NotificationServer) UnregisterPushToken(ctx context.Context, req *chatv1.UnregisterPushTokenRequest) (*chatv1.UnregisterPushTokenResponse, error) {
	return &chatv1.UnregisterPushTokenResponse{}, s.UC.UnregisterPushToken(ctx, userIDFrom(ctx), req.Token)
}
