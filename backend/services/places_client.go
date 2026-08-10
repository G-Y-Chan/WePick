package services

import (
	"backend/util"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"
)

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

// FetchCards executes the text search within a bounding box viewport and maps the results directly to cards
func (pc *PlacesClient) FetchCards(filters util.Filters, pageToken string) ([]util.Card, string, error) {
	// Build the HTTP request with the rectangular viewport constraint
	req, err := pc.buildTextSearchRequest(filters, pageToken)
	if err != nil {
		return nil, "", fmt.Errorf("failed to build request: %w", err)
	}

	// Send the request to Google
	resp, err := pc.HTTPClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("places api network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("google api returned status code: %d", resp.StatusCode)
	}

	// Decode the raw response payload
	var searchResp TextSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, "", fmt.Errorf("failed to decode google response: %w", err)
	}

	// Map every returned place directly into UI deck cards
	cards, hitEdge := pc.mapPlacesToCards(searchResp.Places, filters.Latitude, filters.Longitude, float64(filters.Radius))

	token := searchResp.NextPageToken
	if hitEdge {
		token = "" // Force pagination to stop
	}

	return cards, token, nil
}

func (pc *PlacesClient) buildTextSearchRequest(filters util.Filters, pageToken string) (*http.Request, error) {
	// Convert center coordinates and radius filter into the mandatory rectangular bounding box
	// using the latitude and longitude directly from the filters struct
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
	googleURL := fmt.Sprintf(
		photoEndpoint,
		photoName,
		pc.APIKey,
	)

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

	// Decode the JSON containing the safe CDN URL
	var mediaResp PhotoMediaResponse
	if err := json.NewDecoder(resp.Body).Decode(&mediaResp); err != nil {
		return "", fmt.Errorf("failed to decode google image response: %w", err)
	}

	return mediaResp.PhotoUri, nil
}

// ==========================================
// Geometry & Mapping Helpers
// ==========================================

func calculateBoundingBox(lat, lng float64, radiusInMeters float64) Viewport {
	const earthRadius = 6378137.0 // Earth's equatorial radius in meters

	// Angular offsets in radians
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
	// These must match Google's official Place Types exactly
	switch category {
	case "cafe":
		return "cafe"
	case "bar":
		return "bar"
	default:
		return "restaurant"
	}
}

func (pc *PlacesClient) mapPlacesToCards(places []Place, centerLat, centerLng float64, maxRadius float64) ([]util.Card, bool) {
	cards := make([]util.Card, 0, len(places))
	hitEdge := false

	for _, place := range places {
		distance := haversineDistance(centerLat, centerLng, place.Location.Latitude, place.Location.Longitude)

		if distance > maxRadius {
			hitEdge = true
			break
		}

		priceDisplay := mapPriceLevel(place.PriceLevel)

		photoName := ""
		photoURL := ""
		if len(place.Photos) > 0 {
			photoName = place.Photos[0].Name
			if resolved, err := pc.GetPhotoURL(photoName); err == nil {
				photoURL = resolved
			}
		}

		cards = append(cards, util.Card{
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
			PhotoURL:    photoURL,
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
	const R = 6371000.0 // Earth's radius in meters
	phi1 := lat1 * math.Pi / 180.0
	phi2 := lat2 * math.Pi / 180.0
	deltaPhi := (lat2 - lat1) * math.Pi / 180.0
	deltaLambda := (lon2 - lon1) * math.Pi / 180.0

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
