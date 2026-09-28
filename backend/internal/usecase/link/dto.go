package link

import "github.com/newt239/chat/internal/domain/entity"

type FetchOGPInput struct {
	URL string
}

type FetchOGPOutput struct {
	OGPData entity.OGPData
}
