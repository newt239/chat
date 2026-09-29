package rpc

import (
	"context"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	draftuc "github.com/newt239/chat/internal/usecase/draft"
)

type DraftServer struct {
	UC *draftuc.Interactor
}

func draftTarget(ctx context.Context, channelID string, parentID *string) domainrepository.DraftTarget {
	return domainrepository.DraftTarget{UserID: userIDFrom(ctx), ChannelID: channelID, ParentID: parentID}
}

func (s *DraftServer) SaveDraft(ctx context.Context, req *chatv1.SaveDraftRequest) (*chatv1.SaveDraftResponse, error) {
	d, err := s.UC.Save(ctx, draftuc.SaveInput{Target: draftTarget(ctx, req.ChannelId, req.ParentId), Body: req.Body})
	if err != nil {
		return nil, err
	}
	return &chatv1.SaveDraftResponse{Draft: presenter.OptionalDraft(d)}, nil
}

func (s *DraftServer) GetDraft(ctx context.Context, req *chatv1.GetDraftRequest) (*chatv1.GetDraftResponse, error) {
	d, err := s.UC.Get(ctx, draftTarget(ctx, req.ChannelId, req.ParentId))
	if err != nil {
		return nil, err
	}
	return &chatv1.GetDraftResponse{Draft: presenter.OptionalDraft(d)}, nil
}

func (s *DraftServer) DeleteDraft(ctx context.Context, req *chatv1.DeleteDraftRequest) (*chatv1.DeleteDraftResponse, error) {
	if err := s.UC.Delete(ctx, draftTarget(ctx, req.ChannelId, req.ParentId)); err != nil {
		return nil, err
	}
	return &chatv1.DeleteDraftResponse{}, nil
}

func (s *DraftServer) ListDrafts(ctx context.Context, req *chatv1.ListDraftsRequest) (*chatv1.ListDraftsResponse, error) {
	drafts, err := s.UC.List(ctx, userIDFrom(ctx), req.WorkspaceId)
	if err != nil {
		return nil, err
	}
	return &chatv1.ListDraftsResponse{Drafts: presenter.ConvertAll(drafts, presenter.Draft)}, nil
}
