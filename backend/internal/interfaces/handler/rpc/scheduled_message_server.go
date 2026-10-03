package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	scheduledmessageuc "github.com/newt239/chat/internal/usecase/scheduledmessage"
)

type ScheduledMessageServer struct {
	UC *scheduledmessageuc.Interactor
}

func (s *ScheduledMessageServer) CreateScheduledMessage(ctx context.Context, req *chatv1.CreateScheduledMessageRequest) (*chatv1.CreateScheduledMessageResponse, error) {
	err := s.UC.Schedule(ctx, scheduledmessageuc.ScheduleInput{
		UserID:        userIDFrom(ctx),
		ChannelID:     req.ChannelId,
		ParentID:      req.ParentId,
		Body:          req.Body,
		AttachmentIDs: req.AttachmentIds,
		Location:      locationInput(req.Location),
		ScheduledAt:   req.ScheduledAt.AsTime(),
	})
	return &chatv1.CreateScheduledMessageResponse{}, err
}

func (s *ScheduledMessageServer) ListScheduledMessages(ctx context.Context, req *chatv1.ListScheduledMessagesRequest) (*chatv1.ListScheduledMessagesResponse, error) {
	out, err := s.UC.List(ctx, userIDFrom(ctx), req.WorkspaceId)
	if err != nil {
		return nil, err
	}
	return &chatv1.ListScheduledMessagesResponse{ScheduledMessages: presenter.ConvertAll(out, presenter.ScheduledMessage)}, nil
}

func (s *ScheduledMessageServer) UpdateScheduledMessage(ctx context.Context, req *chatv1.UpdateScheduledMessageRequest) (*chatv1.UpdateScheduledMessageResponse, error) {
	err := s.UC.Reschedule(ctx, scheduledmessageuc.RescheduleInput{
		ID:          req.Id,
		UserID:      userIDFrom(ctx),
		Body:        req.Body,
		ScheduledAt: req.ScheduledAt.AsTime(),
	})
	return &chatv1.UpdateScheduledMessageResponse{}, err
}

func (s *ScheduledMessageServer) DeleteScheduledMessage(ctx context.Context, req *chatv1.DeleteScheduledMessageRequest) (*chatv1.DeleteScheduledMessageResponse, error) {
	return &chatv1.DeleteScheduledMessageResponse{}, s.UC.Delete(ctx, req.Id, userIDFrom(ctx))
}

func (s *ScheduledMessageServer) SendScheduledMessageNow(ctx context.Context, req *chatv1.SendScheduledMessageNowRequest) (*chatv1.SendScheduledMessageNowResponse, error) {
	return &chatv1.SendScheduledMessageNowResponse{}, s.UC.SendNow(ctx, req.Id, userIDFrom(ctx))
}
