package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	realtimeuc "github.com/newt239/chat/internal/usecase/realtime"
)

type RealtimeServer struct {
	UC *realtimeuc.Interactor
}

func (s *RealtimeServer) IssueWebSocketTicket(ctx context.Context, req *chatv1.IssueWebSocketTicketRequest) (*chatv1.IssueWebSocketTicketResponse, error) {
	claims := claimsFrom(ctx)
	ticket, err := s.UC.IssueTicket(ctx, realtimeuc.Ticket{UserID: claims.UserID, SessionID: claims.SessionID, WorkspaceID: req.WorkspaceId})
	if err != nil {
		return nil, err
	}
	return &chatv1.IssueWebSocketTicketResponse{Ticket: ticket}, nil
}
