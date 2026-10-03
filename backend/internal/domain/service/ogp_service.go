package service

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type OGPService interface {
	FetchOGP(ctx context.Context, url string) (*entity.OGPData, error)
}
