package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	exportuc "github.com/newt239/chat/internal/usecase/export"
)

type ExportServer struct {
	UC *exportuc.Interactor
}

func (s *ExportServer) ExportMessages(ctx context.Context, req *chatv1.ExportMessagesRequest) (*chatv1.ExportMessagesResponse, error) {
	out, err := s.UC.ExportMessages(ctx, exportuc.Input{WorkspaceID: req.WorkspaceId, RequesterID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.ExportMessagesResponse{Content: out.Content, FileName: out.FileName}, nil
}
