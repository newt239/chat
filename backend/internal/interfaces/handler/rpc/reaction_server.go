package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	reactionuc "github.com/newt239/chat/internal/usecase/reaction"
)

type ReactionServer struct {
	UC reactionuc.ReactionUseCase
}

func (s *ReactionServer) ListReactions(ctx context.Context, req *chatv1.ListReactionsRequest) (*chatv1.ListReactionsResponse, error) {
	out, err := s.UC.ListReactions(ctx, req.MessageId, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.ListReactionsResponse{Reactions: presenter.ConvertAll(out.Reactions, presenter.Reaction)}, nil
}

func (s *ReactionServer) AddReaction(ctx context.Context, req *chatv1.AddReactionRequest) (*chatv1.AddReactionResponse, error) {
	if err := s.UC.AddReaction(ctx, reactionuc.AddReactionInput{MessageID: req.MessageId, UserID: userIDFrom(ctx), Emoji: req.Emoji}); err != nil {
		return nil, err
	}
	return &chatv1.AddReactionResponse{}, nil
}

func (s *ReactionServer) RemoveReaction(ctx context.Context, req *chatv1.RemoveReactionRequest) (*chatv1.RemoveReactionResponse, error) {
	if err := s.UC.RemoveReaction(ctx, reactionuc.RemoveReactionInput{MessageID: req.MessageId, UserID: userIDFrom(ctx), Emoji: req.Emoji}); err != nil {
		return nil, err
	}
	return &chatv1.RemoveReactionResponse{}, nil
}
