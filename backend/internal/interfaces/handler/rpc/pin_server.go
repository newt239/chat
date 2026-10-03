package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	pinuc "github.com/newt239/chat/internal/usecase/pin"
)

type PinServer struct {
	UC *pinuc.Interactor
}

func (s *PinServer) ListPins(ctx context.Context, req *chatv1.ListPinsRequest) (*chatv1.ListPinsResponse, error) {
	out, err := s.UC.ListPins(ctx, pinuc.ListPinsInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx), Limit: int(req.Limit), Cursor: req.Cursor})
	if err != nil {
		return nil, err
	}
	return &chatv1.ListPinsResponse{Pins: presenter.ConvertAll(out.Pins, presenter.PinnedMessage), NextCursor: out.NextCursor}, nil
}

func (s *PinServer) CreatePin(ctx context.Context, req *chatv1.CreatePinRequest) (*chatv1.CreatePinResponse, error) {
	return &chatv1.CreatePinResponse{}, s.UC.PinMessage(ctx, pinuc.PinInput{ChannelID: req.ChannelId, MessageID: req.MessageId, UserID: userIDFrom(ctx)})
}

func (s *PinServer) DeletePin(ctx context.Context, req *chatv1.DeletePinRequest) (*chatv1.DeletePinResponse, error) {
	return &chatv1.DeletePinResponse{}, s.UC.UnpinMessage(ctx, pinuc.PinInput{ChannelID: req.ChannelId, MessageID: req.MessageId, UserID: userIDFrom(ctx)})
}
