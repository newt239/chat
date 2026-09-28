package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	linkuc "github.com/newt239/chat/internal/usecase/link"
)

type LinkServer struct {
	UC linkuc.LinkUseCase
}

func (s *LinkServer) FetchOgp(ctx context.Context, req *chatv1.FetchOgpRequest) (*chatv1.FetchOgpResponse, error) {
	out, err := s.UC.FetchOGP(ctx, linkuc.FetchOGPInput{URL: req.Url})
	if err != nil {
		return nil, err
	}
	ogp := out.OGPData
	return &chatv1.FetchOgpResponse{Ogp: &chatv1.OgpData{
		Title:       ogp.Title,
		Description: ogp.Description,
		ImageUrl:    ogp.ImageURL,
		SiteName:    ogp.SiteName,
		CardType:    ogp.CardType,
	}}, nil
}
