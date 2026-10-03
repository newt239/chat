package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	reactionuc "github.com/newt239/chat/internal/usecase/reaction"
)

type ReactionServer struct {
	UC *reactionuc.Interactor
}

func (s *ReactionServer) AddReaction(ctx context.Context, req *chatv1.AddReactionRequest) (*chatv1.AddReactionResponse, error) {
	return &chatv1.AddReactionResponse{}, s.UC.AddReaction(ctx, reactionuc.ReactionInput{MessageID: req.MessageId, UserID: userIDFrom(ctx), Emoji: req.Emoji})
}

func (s *ReactionServer) RemoveReaction(ctx context.Context, req *chatv1.RemoveReactionRequest) (*chatv1.RemoveReactionResponse, error) {
	return &chatv1.RemoveReactionResponse{}, s.UC.RemoveReaction(ctx, reactionuc.ReactionInput{MessageID: req.MessageId, UserID: userIDFrom(ctx), Emoji: req.Emoji})
}
