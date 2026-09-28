package rpc

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

const defaultMessageLimit = 20

type MessageServer struct {
	UC messageuc.MessageUseCase
}

func listMessagesInput(ctx context.Context, channelID string, limit int32, since, until *timestamppb.Timestamp) messageuc.ListMessagesInput {
	input := messageuc.ListMessagesInput{ChannelID: channelID, UserID: userIDFrom(ctx), Limit: defaultMessageLimit}
	if limit > 0 {
		input.Limit = int(limit)
	}
	input.Since = optionalTime(since)
	input.Until = optionalTime(until)
	return input
}

func optionalTime(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	converted := t.AsTime()
	return &converted
}

func (s *MessageServer) ListMessages(ctx context.Context, req *chatv1.ListMessagesRequest) (*chatv1.ListMessagesResponse, error) {
	out, err := s.UC.ListMessages(ctx, listMessagesInput(ctx, req.ChannelId, req.Limit, req.Since, req.Until))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListMessagesResponse{Messages: presenter.ConvertAll(out.Messages, presenter.TimelineItem), HasMore: out.HasMore}, nil
}

func (s *MessageServer) ListMessagesWithThread(ctx context.Context, req *chatv1.ListMessagesWithThreadRequest) (*chatv1.ListMessagesWithThreadResponse, error) {
	input := listMessagesInput(ctx, req.ChannelId, req.Limit, req.Since, req.Until)
	// has_more はスレッド付きの一覧では求まらないため通常の一覧から得る
	list, err := s.UC.ListMessages(ctx, input)
	if err != nil {
		return nil, err
	}
	out, err := s.UC.ListMessagesWithThread(ctx, input)
	if err != nil {
		return nil, err
	}
	return &chatv1.ListMessagesWithThreadResponse{Messages: presenter.ConvertAll(out, presenter.MessageWithThread), HasMore: list.HasMore}, nil
}

func (s *MessageServer) CreateMessage(ctx context.Context, req *chatv1.CreateMessageRequest) (*chatv1.CreateMessageResponse, error) {
	out, err := s.UC.CreateMessage(ctx, messageuc.CreateMessageInput{
		ChannelID:     req.ChannelId,
		UserID:        userIDFrom(ctx),
		Body:          req.Body,
		ParentID:      req.ParentId,
		AttachmentIDs: req.AttachmentIds,
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.CreateMessageResponse{Message: presenter.Message(*out)}, nil
}

func (s *MessageServer) UpdateMessage(ctx context.Context, req *chatv1.UpdateMessageRequest) (*chatv1.UpdateMessageResponse, error) {
	out, err := s.UC.UpdateMessage(ctx, messageuc.UpdateMessageInput{MessageID: req.MessageId, EditorID: userIDFrom(ctx), Body: req.Body})
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdateMessageResponse{Message: presenter.Message(*out)}, nil
}

func (s *MessageServer) DeleteMessage(ctx context.Context, req *chatv1.DeleteMessageRequest) (*chatv1.DeleteMessageResponse, error) {
	if err := s.UC.DeleteMessage(ctx, messageuc.DeleteMessageInput{MessageID: req.MessageId, ExecutorID: userIDFrom(ctx)}); err != nil {
		return nil, err
	}
	return &chatv1.DeleteMessageResponse{}, nil
}
