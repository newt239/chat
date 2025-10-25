package service

import "context"

// OGPData represents Open Graph Protocol data
type OGPData struct {
	Title       *string
	Description *string
	ImageURL    *string
	SiteName    *string
	CardType    *string
}

// OGPService defines the interface for OGP operations
type OGPService interface {
	FetchOGP(ctx context.Context, url string) (*OGPData, error)
	ExtractURLs(text string) []string
}
