package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	webhookuc "github.com/newt239/chat/internal/usecase/webhook"
)

type WebhookServer struct {
	UC *webhookuc.Interactor
}

func (s *WebhookServer) ListWebhooks(ctx context.Context, req *chatv1.ListWebhooksRequest) (*chatv1.ListWebhooksResponse, error) {
	out, err := s.UC.List(ctx, webhookuc.ListInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.ListWebhooksResponse{Webhooks: presenter.ConvertAll(out, presenter.Webhook)}, nil
}

func (s *WebhookServer) CreateWebhook(ctx context.Context, req *chatv1.CreateWebhookRequest) (*chatv1.CreateWebhookResponse, error) {
	out, err := s.UC.Create(ctx, webhookuc.CreateInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx), Name: req.Name, AvatarURL: req.AvatarUrl})
	if err != nil {
		return nil, err
	}
	return &chatv1.CreateWebhookResponse{Webhook: presenter.Webhook(out.Webhook), Token: out.Token}, nil
}

func (s *WebhookServer) UpdateWebhook(ctx context.Context, req *chatv1.UpdateWebhookRequest) (*chatv1.UpdateWebhookResponse, error) {
	out, err := s.UC.Update(ctx, webhookuc.UpdateInput{WebhookID: req.WebhookId, UserID: userIDFrom(ctx), Name: req.Name, AvatarURL: req.AvatarUrl})
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdateWebhookResponse{Webhook: presenter.Webhook(*out)}, nil
}

func (s *WebhookServer) RegenerateWebhookToken(ctx context.Context, req *chatv1.RegenerateWebhookTokenRequest) (*chatv1.RegenerateWebhookTokenResponse, error) {
	token, err := s.UC.RegenerateToken(ctx, webhookuc.TargetInput{WebhookID: req.WebhookId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.RegenerateWebhookTokenResponse{Token: token}, nil
}

func (s *WebhookServer) DeleteWebhook(ctx context.Context, req *chatv1.DeleteWebhookRequest) (*chatv1.DeleteWebhookResponse, error) {
	if err := s.UC.Delete(ctx, webhookuc.TargetInput{WebhookID: req.WebhookId, UserID: userIDFrom(ctx)}); err != nil {
		return nil, err
	}
	return &chatv1.DeleteWebhookResponse{}, nil
}
