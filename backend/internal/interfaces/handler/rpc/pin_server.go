package rpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	pinuc "github.com/newt239/chat/internal/usecase/pin"
)

const defaultPinLimit = 100

type PinServer struct {
	UC pinuc.PinUseCase
}

func (s *PinServer) ListPins(ctx context.Context, req *chatv1.ListPinsRequest) (*chatv1.ListPinsResponse, error) {
	input := pinuc.ListPinsInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx), Limit: defaultPinLimit, Cursor: req.Cursor}
	if req.Limit > 0 {
		input.Limit = int(req.Limit)
	}
	out, err := s.UC.ListPins(ctx, input)
	if err != nil {
		return nil, err
	}
	return &chatv1.ListPinsResponse{
		Pins: presenter.ConvertAll(out.Pins, func(p pinuc.PinnedMessageOutput) *chatv1.PinnedMessage {
			return &chatv1.PinnedMessage{Message: presenter.Message(p.Message), PinnedBy: p.PinnedBy, PinnedAt: timestamppb.New(p.PinnedAt)}
		}),
		NextCursor: out.NextCursor,
	}, nil
}

func (s *PinServer) CreatePin(ctx context.Context, req *chatv1.CreatePinRequest) (*chatv1.CreatePinResponse, error) {
	if err := s.UC.PinMessage(ctx, pinuc.PinMessageInput{ChannelID: req.ChannelId, MessageID: req.MessageId, UserID: userIDFrom(ctx)}); err != nil {
		return nil, err
	}
	return &chatv1.CreatePinResponse{}, nil
}

func (s *PinServer) DeletePin(ctx context.Context, req *chatv1.DeletePinRequest) (*chatv1.DeletePinResponse, error) {
	if err := s.UC.UnpinMessage(ctx, pinuc.UnpinMessageInput{ChannelID: req.ChannelId, MessageID: req.MessageId, UserID: userIDFrom(ctx)}); err != nil {
		return nil, err
	}
	return &chatv1.DeletePinResponse{}, nil
}
