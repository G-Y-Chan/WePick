package services

// Places API Endpoint for Text Search (New)
const textSearchEndpoint = "https://places.googleapis.com/v1/places:searchText"
const photoEndpoint = "https://places.googleapis.com/v1/%s/media?maxHeightPx=800&maxWidthPx=800&skipHttpRedirect=true&key=%s"

type LatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Viewport struct {
	Low  LatLng `json:"low"`  // Southwest corner
	High LatLng `json:"high"` // Northeast corner
}

type LocationRestriction struct {
	Rectangle Viewport `json:"rectangle"`
}

type TextSearchRequest struct {
	TextQuery           string              `json:"textQuery"`
	PageSize            int                 `json:"pageSize,omitempty"`
	OpenNow             bool                `json:"openNow,omitempty"`
	PageToken           string              `json:"pageToken,omitempty"`
	LocationRestriction LocationRestriction `json:"locationRestriction"`
	IncludedType        string              `json:"includedType,omitempty"`
	StrictTypeFiltering bool                `json:"strictTypeFiltering,omitempty"`
	RankPreference      string              `json:"rankPreference,omitempty"`
}

type LocalizedText struct {
	Text         string `json:"text"`
	LanguageCode string `json:"languageCode"`
}

type OpeningHours struct {
	OpenNow bool `json:"openNow"`
}

type Photo struct {
	Name string `json:"name"`
}

type Place struct {
	ID                     string        `json:"id"`
	DisplayName            LocalizedText `json:"displayName"`
	ShortFormattedAddress  string        `json:"shortFormattedAddress"`
	PrimaryTypeDisplayName LocalizedText `json:"primaryTypeDisplayName"`
	PriceLevel             string        `json:"priceLevel"`
	Rating                 float64       `json:"rating"`
	UserRatingCount        int           `json:"userRatingCount"`
	CurrentOpeningHours    OpeningHours  `json:"currentOpeningHours"`
	EditorialSummary       LocalizedText `json:"editorialSummary"`
	Location               LatLng        `json:"location"`
	Photos                 []Photo       `json:"photos,omitempty"`
}

type TextSearchResponse struct {
	Places        []Place `json:"places"`
	NextPageToken string  `json:"nextPageToken,omitempty"`
}

type PhotoMediaResponse struct {
	PhotoUri string `json:"photoUri"`
}
