package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	customemojiuc "github.com/newt239/chat/internal/usecase/customemoji"
)

type CustomEmojiServer struct {
	UC *customemojiuc.Interactor
}

func (s *CustomEmojiServer) ListCustomEmojis(ctx context.Context, req *chatv1.ListCustomEmojisRequest) (*chatv1.ListCustomEmojisResponse, error) {
	out, err := s.UC.List(ctx, customemojiuc.ListInput{WorkspaceID: req.WorkspaceId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.ListCustomEmojisResponse{Emojis: presenter.ConvertAll(out, presenter.CustomEmoji)}, nil
}

func (s *CustomEmojiServer) PresignCustomEmojiUpload(ctx context.Context, req *chatv1.PresignCustomEmojiUploadRequest) (*chatv1.PresignCustomEmojiUploadResponse, error) {
	out, err := s.UC.Presign(ctx, customemojiuc.PresignInput{
		WorkspaceID: req.WorkspaceId,
		UserID:      userIDFrom(ctx),
		ContentType: req.ContentType,
		SizeBytes:   req.SizeBytes,
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.PresignCustomEmojiUploadResponse{UploadId: out.UploadID, UploadUrl: out.UploadURL}, nil
}

func (s *CustomEmojiServer) CreateCustomEmoji(ctx context.Context, req *chatv1.CreateCustomEmojiRequest) (*chatv1.CreateCustomEmojiResponse, error) {
	out, err := s.UC.Create(ctx, customemojiuc.CreateInput{
		WorkspaceID: req.WorkspaceId,
		UserID:      userIDFrom(ctx),
		Name:        req.Name,
		UploadID:    req.UploadId,
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.CreateCustomEmojiResponse{Emoji: presenter.CustomEmoji(*out)}, nil
}

func (s *CustomEmojiServer) DeleteCustomEmoji(ctx context.Context, req *chatv1.DeleteCustomEmojiRequest) (*chatv1.DeleteCustomEmojiResponse, error) {
	return &chatv1.DeleteCustomEmojiResponse{}, s.UC.Delete(ctx, customemojiuc.DeleteInput{WorkspaceID: req.WorkspaceId, UserID: userIDFrom(ctx), EmojiID: req.EmojiId})
}
