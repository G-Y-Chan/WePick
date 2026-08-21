package places

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"time"

	"backend/internal/room"
)

const textSearchEndpoint = "https://places.googleapis.com/v1/places:searchText"
const photoEndpoint = "https://places.googleapis.com/v1/%s/media?maxHeightPx=4800&maxWidthPx=4800&skipHttpRedirect=true&key=%s"

var photoNameRe = regexp.MustCompile(`^places/[A-Za-z0-9_-]+/photos/[A-Za-z0-9_-]+$`)

type LatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Viewport struct {
	Low  LatLng `json:"low"`
	High LatLng `json:"high"`
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
	Name     string `json:"name"`
	WidthPx  int    `json:"widthPx,omitempty"`
	HeightPx int    `json:"heightPx,omitempty"`
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

type PlacesClient struct {
	APIKey     string
	HTTPClient *http.Client
}

func NewPlacesClient(apiKey string) *PlacesClient {
	return &PlacesClient{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (pc *PlacesClient) FetchCards(filters room.Filters, pageToken string) ([]room.Card, string, error) {
	req, err := pc.buildTextSearchRequest(filters, pageToken)
	if err != nil {
		return nil, "", fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := pc.HTTPClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("places api network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("google api returned status code: %d", resp.StatusCode)
	}

	var searchResp TextSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, "", fmt.Errorf("failed to decode google response: %w", err)
	}

	cards, hitEdge := mapPlacesToCards(searchResp.Places, filters.Latitude, filters.Longitude, float64(filters.Radius))

	token := searchResp.NextPageToken
	if hitEdge {
		token = ""
	}

	return cards, token, nil
}

func (pc *PlacesClient) buildTextSearchRequest(filters room.Filters, pageToken string) (*http.Request, error) {
	boundingBox := calculateBoundingBox(filters.Latitude, filters.Longitude, float64(filters.Radius))

	reqBody := TextSearchRequest{
		TextQuery: mapCategoryToTextQuery(filters.Category),
		PageSize:  20,
		OpenNow:   filters.OpenNow,
		PageToken: pageToken,
		LocationRestriction: LocationRestriction{
			Rectangle: boundingBox,
		},
		IncludedType:        mapCategoryToIncludedType(filters.Category),
		StrictTypeFiltering: true,
		RankPreference:      "DISTANCE",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, textSearchEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", pc.APIKey)
	req.Header.Set("X-Goog-FieldMask", "places.id,places.displayName.text,places.shortFormattedAddress,places.primaryTypeDisplayName.text,places.priceLevel,places.rating,places.userRatingCount,places.currentOpeningHours.openNow,places.editorialSummary.text,places.location,places.photos,nextPageToken")

	return req, nil
}

func (pc *PlacesClient) GetPhotoURL(photoName string) (string, error) {
	if !photoNameRe.MatchString(photoName) {
		return "", fmt.Errorf("invalid photo name format: %q", photoName)
	}

	googleURL := fmt.Sprintf(photoEndpoint, photoName, pc.APIKey)

	req, err := http.NewRequest(http.MethodGet, googleURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := pc.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("places api network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google api returned status code: %d", resp.StatusCode)
	}

	var mediaResp PhotoMediaResponse
	if err := json.NewDecoder(resp.Body).Decode(&mediaResp); err != nil {
		return "", fmt.Errorf("failed to decode google image response: %w", err)
	}

	return mediaResp.PhotoUri, nil
}

func calculateBoundingBox(lat, lng float64, radiusInMeters float64) Viewport {
	const earthRadius = 6378137.0
	dLat := radiusInMeters / earthRadius
	dLng := radiusInMeters / (earthRadius * math.Cos(lat*math.Pi/180.0))

	return Viewport{
		Low: LatLng{
			Latitude:  lat - (dLat * 180.0 / math.Pi),
			Longitude: lng - (dLng * 180.0 / math.Pi),
		},
		High: LatLng{
			Latitude:  lat + (dLat * 180.0 / math.Pi),
			Longitude: lng + (dLng * 180.0 / math.Pi),
		},
	}
}

func mapCategoryToTextQuery(category string) string {
	switch category {
	case "cafe":
		return "cafes and coffee shops"
	case "bar":
		return "bars and pubs"
	default:
		return "restaurants"
	}
}

func mapCategoryToIncludedType(category string) string {
	switch category {
	case "cafe":
		return "cafe"
	case "bar":
		return "bar"
	default:
		return "restaurant"
	}
}

func mapPlacesToCards(places []Place, centerLat, centerLng float64, maxRadius float64) ([]room.Card, bool) {
	cards := make([]room.Card, 0, len(places))
	hitEdge := false

	for _, place := range places {
		distance := haversineDistance(centerLat, centerLng, place.Location.Latitude, place.Location.Longitude)
		if distance > maxRadius {
			hitEdge = true
			break
		}

		priceDisplay := mapPriceLevel(place.PriceLevel)

		photoName := ""
		if len(place.Photos) > 0 {
			photoName = place.Photos[0].Name
		}

		cards = append(cards, room.Card{
			ID:          place.ID,
			Title:       place.DisplayName.Text,
			Category:    place.PrimaryTypeDisplayName.Text,
			PriceLevel:  priceDisplay,
			Rating:      place.Rating,
			ReviewCount: place.UserRatingCount,
			OpenNow:     place.CurrentOpeningHours.OpenNow,
			Summary:     place.EditorialSummary.Text,
			Address:     place.ShortFormattedAddress,
			PhotoName:   photoName,
		})
	}
	return cards, hitEdge
}

func mapPriceLevel(level string) string {
	switch level {
	case "PRICE_LEVEL_FREE":
		return "Free"
	case "PRICE_LEVEL_INEXPENSIVE":
		return "$"
	case "PRICE_LEVEL_MODERATE":
		return "$$"
	case "PRICE_LEVEL_EXPENSIVE":
		return "$$$"
	case "PRICE_LEVEL_VERY_EXPENSIVE":
		return "$$$$"
	default:
		return "Price level unknown"
	}
}

func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000.0
	phi1 := lat1 * math.Pi / 180.0
	phi2 := lat2 * math.Pi / 180.0
	deltaPhi := (lat2 - lat1) * math.Pi / 180.0
	deltaLambda := (lon2 - lon1) * math.Pi / 180.0

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}