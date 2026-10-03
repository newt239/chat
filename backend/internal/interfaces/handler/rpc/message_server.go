package rpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

type MessageServer struct {
	UC *messageuc.Interactor
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
	return messageuc.ListMessagesInput{
		ChannelID:          req.GetChannelId(),
		UserID:             userIDFrom(ctx),
		Limit:              int(req.GetLimit()),
		Since:              optionalTime(req.GetSince()),
		Until:              optionalTime(req.GetUntil()),
		IncludeDescendants: req.GetIncludeDescendants(),
	}
}

func (s *MessageServer) ListMessages(ctx context.Context, req *chatv1.ListMessagesRequest) (*chatv1.ListMessagesResponse, error) {
	input := listMessagesInput(ctx, req)
	input.Around = optionalTime(req.GetAround())
	out, err := s.UC.ListMessages(ctx, input)
	if err != nil {
		return nil, err
	}
	return &chatv1.ListMessagesResponse{Messages: presenter.ConvertAll(out.Messages, presenter.TimelineItem), HasMore: out.HasMore, HasNewer: out.HasNewer}, nil
}

func (s *MessageServer) ListMessagesWithThread(ctx context.Context, req *chatv1.ListMessagesWithThreadRequest) (*chatv1.ListMessagesWithThreadResponse, error) {
	out, err := s.UC.ListMessagesWithThread(ctx, listMessagesInput(ctx, req))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListMessagesWithThreadResponse{Messages: presenter.ConvertAll(out.Messages, presenter.MessageWithThread), HasMore: out.HasMore}, nil
}

func (s *MessageServer) CreateMessage(ctx context.Context, req *chatv1.CreateMessageRequest) (*chatv1.CreateMessageResponse, error) {
	out, err := s.UC.CreateMessage(ctx, messageuc.CreateMessageInput{
		ChannelID:     req.ChannelId,
		UserID:        userIDFrom(ctx),
		Body:          req.Body,
		ParentID:      req.ParentId,
		AttachmentIDs: req.AttachmentIds,
		Location:      locationInput(req.Location),
		Poll:          pollInput(req.Poll),
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
	return &chatv1.DeleteMessageResponse{}, s.UC.DeleteMessage(ctx, messageuc.MessageInput{MessageID: req.MessageId, UserID: userIDFrom(ctx)})
}

func (s *MessageServer) GetMessagePreview(ctx context.Context, req *chatv1.GetMessagePreviewRequest) (*chatv1.GetMessagePreviewResponse, error) {
	out, err := s.UC.GetMessagePreview(ctx, messageuc.MessageInput{MessageID: req.MessageId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.GetMessagePreviewResponse{Preview: presenter.MessagePreview(out)}, nil
}

// pollInput はリクエストの投票を読み替えます。値の範囲は protovalidate で検証済み
func pollInput(p *chatv1.PollInput) *messageuc.PollInput {
	if p == nil {
		return nil
	}
	input := &messageuc.PollInput{
		Question:      p.Question,
		Mode:          fromProto(presenter.PollModes, p.Mode),
		AllowMultiple: p.AllowMultiple,
		Anonymous:     p.Anonymous,
		ClosesAt:      optionalTime(p.ClosesAt),
	}
	for _, o := range p.Options {
		input.Options = append(input.Options, messageuc.PollOptionInput{Label: o.Label, StartsAt: optionalTime(o.StartsAt), AllDay: o.AllDay})
	}
	return input
}
