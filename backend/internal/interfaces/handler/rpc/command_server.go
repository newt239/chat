package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	commanduc "github.com/newt239/chat/internal/usecase/command"
)

type CommandServer struct {
	UC *commanduc.Interactor
}

func (s *CommandServer) ExecuteCommand(ctx context.Context, req *chatv1.ExecuteCommandRequest) (*chatv1.ExecuteCommandResponse, error) {
	out, err := s.UC.Execute(ctx, commanduc.ExecuteInput{UserID: userIDFrom(ctx), ChannelID: req.ChannelId, ParentID: req.ParentId, Text: req.Text})
	if err != nil {
		return nil, err
	}
	return &chatv1.ExecuteCommandResponse{Message: presenter.Message(*out)}, nil
}
