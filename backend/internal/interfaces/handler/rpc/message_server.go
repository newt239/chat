package rpc

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

const defaultMessageLimit = 20

type MessageServer struct {
	UC messageuc.MessageUseCase
}

// listMessagesRequest は ListMessages と ListMessagesWithThread のリクエストに共通する項目です
type listMessagesRequest interface {
	GetChannelId() string
	GetLimit() int32
	GetSince() *timestamppb.Timestamp
	GetUntil() *timestamppb.Timestamp
	GetIncludeDescendants() bool
}

func listMessagesInput(ctx context.Context, req listMessagesRequest) messageuc.ListMessagesInput {
	input := messageuc.ListMessagesInput{
		ChannelID:          req.GetChannelId(),
		UserID:             userIDFrom(ctx),
		Limit:              defaultMessageLimit,
		Since:              optionalTime(req.GetSince()),
		Until:              optionalTime(req.GetUntil()),
		IncludeDescendants: req.GetIncludeDescendants(),
	}
	if req.GetLimit() > 0 {
		input.Limit = int(req.GetLimit())
	}
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
	out, err := s.UC.ListMessages(ctx, listMessagesInput(ctx, req))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListMessagesResponse{Messages: presenter.ConvertAll(out.Messages, presenter.TimelineItem), HasMore: out.HasMore}, nil
}

func (s *MessageServer) ListMessagesWithThread(ctx context.Context, req *chatv1.ListMessagesWithThreadRequest) (*chatv1.ListMessagesWithThreadResponse, error) {
	input := listMessagesInput(ctx, req)
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
		Location:      locationInput(req.Location),
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

func (s *MessageServer) GetMessagePreview(ctx context.Context, req *chatv1.GetMessagePreviewRequest) (*chatv1.GetMessagePreviewResponse, error) {
	out, err := s.UC.GetMessagePreview(ctx, messageuc.GetMessagePreviewInput{MessageID: req.MessageId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.GetMessagePreviewResponse{Preview: presenter.MessagePreview(*out)}, nil
}

func locationInput(l *chatv1.MessageLocation) *entity.MessageLocation {
	if l == nil {
		return nil
	}
	return &entity.MessageLocation{Latitude: l.Latitude, Longitude: l.Longitude, AccuracyMeters: l.AccuracyMeters, Label: l.Label}
}
