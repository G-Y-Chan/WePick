package places

// =============================================================================
// Google Places API — Wire DTOs (package-private)
// =============================================================================
// These types mirror the JSON shape of Google Places API (New) v1 responses and
// request bodies. They are deliberately unexported so nothing outside this
// package can couple to Google's wire schema — the anti-corruption layer in
// mapper.go is the sole bridge to the domain model.

const textSearchEndpoint = "https://places.googleapis.com/v1/places:searchText"
const photoEndpoint = "https://places.googleapis.com/v1/%s/media?maxHeightPx=4800&maxWidthPx=4800&skipHttpRedirect=true&key=%s"

type latLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type viewport struct {
	Low  latLng `json:"low"`  // Southwest corner
	High latLng `json:"high"` // Northeast corner
}

type locationRestriction struct {
	Rectangle viewport `json:"rectangle"`
}

type textSearchRequest struct {
	TextQuery           string              `json:"textQuery"`
	PageSize            int                 `json:"pageSize,omitempty"`
	OpenNow             bool                `json:"openNow,omitempty"`
	PageToken           string              `json:"pageToken,omitempty"`
	LocationRestriction locationRestriction `json:"locationRestriction"`
	IncludedType        string              `json:"includedType,omitempty"`
	StrictTypeFiltering bool                `json:"strictTypeFiltering,omitempty"`
	RankPreference      string              `json:"rankPreference,omitempty"`
}

type localizedText struct {
	Text         string `json:"text"`
	LanguageCode string `json:"languageCode"`
}

type openingHours struct {
	OpenNow bool `json:"openNow"`
}

type photo struct {
	Name     string `json:"name"`
	WidthPx  int    `json:"widthPx,omitempty"`
	HeightPx int    `json:"heightPx,omitempty"`
}

type place struct {
	ID                     string        `json:"id"`
	DisplayName            localizedText `json:"displayName"`
	ShortFormattedAddress  string        `json:"shortFormattedAddress"`
	PrimaryTypeDisplayName localizedText `json:"primaryTypeDisplayName"`
	PriceLevel             string        `json:"priceLevel"`
	Rating                 float64       `json:"rating"`
	UserRatingCount        int           `json:"userRatingCount"`
	CurrentOpeningHours    openingHours  `json:"currentOpeningHours"`
	EditorialSummary       localizedText `json:"editorialSummary"`
	Location               latLng        `json:"location"`
	Photos                 []photo       `json:"photos,omitempty"`
}

type textSearchResponse struct {
	Places        []place `json:"places"`
	NextPageToken string  `json:"nextPageToken,omitempty"`
}

type photoMediaResponse struct {
	PhotoUri string `json:"photoUri"`
}
