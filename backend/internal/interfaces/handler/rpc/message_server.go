package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

type MessageServer struct {
	UC *messageuc.Interactor
}

func (s *MessageServer) ListMessages(ctx context.Context, req *chatv1.ListMessagesRequest) (*chatv1.ListMessagesResponse, error) {
	out, err := s.UC.ListMessages(ctx, messageuc.ListMessagesInput{
		ChannelID:          req.ChannelId,
		UserID:             userIDFrom(ctx),
		Limit:              int(req.Limit),
		Since:              optionalTime(req.Since),
		Until:              optionalTime(req.Until),
		Around:             optionalTime(req.Around),
		IncludeDescendants: req.IncludeDescendants,
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.ListMessagesResponse{Messages: presenter.ConvertAll(out.Messages, presenter.TimelineItem), HasMore: out.HasMore, HasNewer: out.HasNewer}, nil
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
