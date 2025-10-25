package link

import (
	"context"
	"fmt"

	"github.com/newt239/chat/internal/domain/service"
)

type LinkUseCase interface {
	FetchOGP(ctx context.Context, input FetchOGPInput) (*FetchOGPOutput, error)
}

type linkInteractor struct {
	ogpService service.OGPService
}

func NewLinkInteractor(ogpService service.OGPService) LinkUseCase {
	return &linkInteractor{
		ogpService: ogpService,
	}
}

func (i *linkInteractor) FetchOGP(ctx context.Context, input FetchOGPInput) (*FetchOGPOutput, error) {
	ogpData, err := i.ogpService.FetchOGP(ctx, input.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OGP data: %w", err)
	}

	output := &FetchOGPOutput{
		OGPData: OGPData{
			Title:       ogpData.Title,
			Description: ogpData.Description,
			ImageURL:    ogpData.ImageURL,
			SiteName:    ogpData.SiteName,
			CardType:    ogpData.CardType,
		},
	}

	return output, nil
}
