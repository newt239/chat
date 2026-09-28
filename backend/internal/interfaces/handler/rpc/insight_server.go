package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	insightuc "github.com/newt239/chat/internal/usecase/insight"
)

type InsightServer struct {
	UC *insightuc.Interactor
}

func (s *InsightServer) GetInsights(ctx context.Context, req *chatv1.GetInsightsRequest) (*chatv1.GetInsightsResponse, error) {
	out, err := s.UC.GetInsights(ctx, insightuc.Input{WorkspaceID: req.WorkspaceId, RequesterID: userIDFrom(ctx), TimeZone: req.TimeZone})
	if err != nil {
		return nil, err
	}
	return presenter.Insights(out), nil
}
