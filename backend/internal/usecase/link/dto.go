package link

// Input DTOs

type FetchOGPInput struct {
	URL string
}

// Output DTOs

type OGPData struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	ImageURL    *string `json:"imageUrl"`
	SiteName    *string `json:"siteName"`
	CardType    *string `json:"cardType"`
}

type FetchOGPOutput struct {
	OGPData OGPData `json:"ogpData"`
}
