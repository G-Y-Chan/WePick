package places

import (
	"backend/internal/domain/room"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Compile-time contract check: *Client must satisfy room.PlacesProvider.
var _ room.PlacesProvider = (*Client)(nil)

// Client implements room.PlacesProvider by calling the Google Places API (New).
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// New creates a Client backed by the supplied API key and HTTP transport.
// If httpClient is nil a default client with a 10 s timeout is used, matching
// the existing production behaviour.
func New(apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}
	return &Client{
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

// Search executes a Google Places Text Search constrained by the supplied
// filters and optional page token.  It returns the domain-mapped cards and the
// next-page token (empty when no further pages exist or the radius boundary
// was hit).
func (c *Client) Search(ctx context.Context, filters room.SearchFilters, pageToken string) ([]room.Card, string, error) {
	req, err := c.buildTextSearchRequest(filters, pageToken)
	if err != nil {
		return nil, "", fmt.Errorf("failed to build request: %w", err)
	}

	// Attach the caller's context to the outbound HTTP request.
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("places api network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("google api returned status code: %d", resp.StatusCode)
	}

	var searchResp textSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, "", fmt.Errorf("failed to decode google response: %w", err)
	}

	cards, hitEdge := mapPlacesToCards(searchResp.Places, filters.Latitude, filters.Longitude, float64(filters.RadiusM))

	token := searchResp.NextPageToken
	if hitEdge {
		token = "" // Force pagination to stop — edge hit means we're beyond the circle.
	}

	return cards, token, nil
}

// PhotoURL resolves a Google Places photo reference (the "name" resource) into
// a CDN-backed image URL suitable for direct client consumption.
func (c *Client) PhotoURL(ctx context.Context, photoRef string) (string, error) {
	googleURL := fmt.Sprintf(
		photoEndpoint,
		photoRef,
		c.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("places api network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google api returned status code: %d", resp.StatusCode)
	}

	var mediaResp photoMediaResponse
	if err := json.NewDecoder(resp.Body).Decode(&mediaResp); err != nil {
		return "", fmt.Errorf("failed to decode google image response: %w", err)
	}

	return mediaResp.PhotoUri, nil
}

// buildTextSearchRequest constructs an HTTP request for the Google Places Text
// Search endpoint, converting domain filters into the Google wire format.
func (c *Client) buildTextSearchRequest(filters room.SearchFilters, pageToken string) (*http.Request, error) {
	boundingBox := calculateBoundingBox(filters.Latitude, filters.Longitude, float64(filters.RadiusM))

	categoryStr := string(filters.Category)

	reqBody := textSearchRequest{
		TextQuery: mapCategoryToTextQuery(categoryStr),
		PageSize:  20,
		OpenNow:   filters.OpenNow,
		PageToken: pageToken,
		LocationRestriction: locationRestriction{
			Rectangle: boundingBox,
		},
		IncludedType:        mapCategoryToIncludedType(categoryStr),
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
	req.Header.Set("X-Goog-Api-Key", c.apiKey)
	req.Header.Set("X-Goog-FieldMask",
		"places.id,places.displayName.text,places.shortFormattedAddress,places.primaryTypeDisplayName.text,"+
			"places.priceLevel,places.rating,places.userRatingCount,places.currentOpeningHours.openNow,"+
			"places.editorialSummary.text,places.location,places.photos,nextPageToken")

	return req, nil
}
