package rpc

import (
	"context"
	"strings"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	searchuc "github.com/newt239/chat/internal/usecase/search"
)

type SearchServer struct {
	UC searchuc.SearchUseCase
}

func (s *SearchServer) SearchWorkspace(ctx context.Context, req *chatv1.SearchWorkspaceRequest) (*chatv1.SearchWorkspaceResponse, error) {
	filter := strings.ToLower(strings.TrimPrefix(req.Filter.String(), "SEARCH_FILTER_"))
	out, err := s.UC.SearchWorkspace(ctx, searchuc.WorkspaceSearchInput{
		WorkspaceID: req.WorkspaceId,
		RequesterID: userIDFrom(ctx),
		Query:       req.Query,
		Filter:      searchuc.SearchFilter(filter).Normalize(),
		Page:        max(int(req.Page), 1),
		PerPage:     int(req.PerPage),
	})
	if err != nil {
		return nil, err
	}
	return presenter.SearchResult(out), nil
}
