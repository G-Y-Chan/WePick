package places

import (
	"testing"
)

// =============================================================================
// Mapper Tests — table-driven, fixture-based, covering every PriceLevel branch
// and edge cases in mapPlacesToCards.
// =============================================================================

func TestMapPriceLevel(t *testing.T) {
	tests := []struct {
		level string
		want  string
	}{
		{"PRICE_LEVEL_FREE", "Free"},
		{"PRICE_LEVEL_INEXPENSIVE", "$"},
		{"PRICE_LEVEL_MODERATE", "$$"},
		{"PRICE_LEVEL_EXPENSIVE", "$$$"},
		{"PRICE_LEVEL_VERY_EXPENSIVE", "$$$$"},
		{"", "Price level unknown"},
		{"SOME_UNKNOWN_VALUE", "Price level unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			got := mapPriceLevel(tt.level)
			if got != tt.want {
				t.Errorf("mapPriceLevel(%q) = %q, want %q", tt.level, got, tt.want)
			}
		})
	}
}

func TestMapCategoryToTextQuery(t *testing.T) {
	tests := []struct {
		category string
		want     string
	}{
		{"cafe", "cafes and coffee shops"},
		{"bar", "bars and pubs"},
		{"restaurant", "restaurants"},
		{"", "restaurants"},
		{"pizza", "restaurants"}, // unknown → default
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			got := mapCategoryToTextQuery(tt.category)
			if got != tt.want {
				t.Errorf("mapCategoryToTextQuery(%q) = %q, want %q", tt.category, got, tt.want)
			}
		})
	}
}

func TestMapCategoryToIncludedType(t *testing.T) {
	tests := []struct {
		category string
		want     string
	}{
		{"cafe", "cafe"},
		{"bar", "bar"},
		{"restaurant", "restaurant"},
		{"", "restaurant"},
		{"unknown", "restaurant"},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			got := mapCategoryToIncludedType(tt.category)
			if got != tt.want {
				t.Errorf("mapCategoryToIncludedType(%q) = %q, want %q", tt.category, got, tt.want)
			}
		})
	}
}

// fullPlaceFixture returns a fully-populated place for use in mapping tests.
func fullPlaceFixture(id string, lat, lng float64, priceLevel string, photos []photo) place {
	return place{
		ID: id,
		DisplayName: localizedText{
			Text:         "Test Place " + id,
			LanguageCode: "en",
		},
		ShortFormattedAddress: "123 Test St, Testville",
		PrimaryTypeDisplayName: localizedText{
			Text:         "Italian Restaurant",
			LanguageCode: "en",
		},
		PriceLevel:      priceLevel,
		Rating:          4.5,
		UserRatingCount: 200,
		CurrentOpeningHours: openingHours{
			OpenNow: true,
		},
		EditorialSummary: localizedText{
			Text:         "A great place for testing.",
			LanguageCode: "en",
		},
		Location: latLng{
			Latitude:  lat,
			Longitude: lng,
		},
		Photos: photos,
	}
}

func TestMapPlacesToCards_FullMapping(t *testing.T) {
	photos := []photo{{Name: "photos/abc123", WidthPx: 800, HeightPx: 600}}
	p := fullPlaceFixture("place_1", 37.7749, -122.4194, "PRICE_LEVEL_MODERATE", photos)

	cards, hitEdge := mapPlacesToCards([]place{p}, 37.7749, -122.4194, 500)

	if hitEdge {
		t.Error("expected hitEdge = false for a place at the centre")
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}

	c := cards[0]
	if c.ID != "place_1" {
		t.Errorf("ID = %q, want %q", c.ID, "place_1")
	}
	if c.Title != "Test Place place_1" {
		t.Errorf("Title = %q, want %q", c.Title, "Test Place place_1")
	}
	if c.Category != "Italian Restaurant" {
		t.Errorf("Category = %q, want %q", c.Category, "Italian Restaurant")
	}
	if c.PriceLevel != "$$" {
		t.Errorf("PriceLevel = %q, want %q", c.PriceLevel, "$$")
	}
	if c.Rating != 4.5 {
		t.Errorf("Rating = %f, want %f", c.Rating, 4.5)
	}
	if c.ReviewCount != 200 {
		t.Errorf("ReviewCount = %d, want %d", c.ReviewCount, 200)
	}
	if !c.OpenNow {
		t.Error("OpenNow should be true")
	}
	if c.Summary != "A great place for testing." {
		t.Errorf("Summary = %q, want %q", c.Summary, "A great place for testing.")
	}
	if c.Address != "123 Test St, Testville" {
		t.Errorf("Address = %q, want %q", c.Address, "123 Test St, Testville")
	}
	if c.PhotoRef != "photos/abc123" {
		t.Errorf("PhotoRef = %q, want %q", c.PhotoRef, "photos/abc123")
	}
}

func TestMapPlacesToCards_HitEdge(t *testing.T) {
	// Place is 1 km away but the radius is only 500 m → should hit the edge.
	p := fullPlaceFixture("far", 37.7749+0.01, -122.4194, "PRICE_LEVEL_INEXPENSIVE", nil)

	cards, hitEdge := mapPlacesToCards([]place{p}, 37.7749, -122.4194, 500)

	if !hitEdge {
		t.Error("expected hitEdge = true for a place outside the radius")
	}
	if len(cards) != 0 {
		t.Errorf("expected 0 cards (break before append), got %d", len(cards))
	}
}

func TestMapPlacesToCards_EdgeDoesNotSkipEarlier(t *testing.T) {
	// Place 1 is within radius; place 2 is outside.  Place 1 should be
	// included, place 2 should trigger hitEdge and stop iteration.
	p1 := fullPlaceFixture("near", 37.7749, -122.4194, "PRICE_LEVEL_FREE", nil)
	p2 := fullPlaceFixture("far", 37.7749+0.01, -122.4194, "PRICE_LEVEL_MODERATE", nil)

	cards, hitEdge := mapPlacesToCards([]place{p1, p2}, 37.7749, -122.4194, 500)

	if !hitEdge {
		t.Error("expected hitEdge = true")
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].ID != "near" {
		t.Errorf("expected 'near' card, got %q", cards[0].ID)
	}
}

func TestMapPlacesToCards_NoPhotos(t *testing.T) {
	p := fullPlaceFixture("nophoto", 37.7749, -122.4194, "PRICE_LEVEL_EXPENSIVE", nil)

	cards, _ := mapPlacesToCards([]place{p}, 37.7749, -122.4194, 500)
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].PhotoRef != "" {
		t.Errorf("PhotoRef = %q, want empty string when no photos", cards[0].PhotoRef)
	}
}

func TestMapPlacesToCards_Empty(t *testing.T) {
	cards, hitEdge := mapPlacesToCards(nil, 37.7749, -122.4194, 500)
	if len(cards) != 0 {
		t.Errorf("expected 0 cards, got %d", len(cards))
	}
	if hitEdge {
		t.Error("expected hitEdge = false for empty input")
	}
}

func TestMapPlacesToCards_AllPriceLevels(t *testing.T) {
	levels := []struct {
		api  string
		want string
	}{
		{"PRICE_LEVEL_FREE", "Free"},
		{"PRICE_LEVEL_INEXPENSIVE", "$"},
		{"PRICE_LEVEL_MODERATE", "$$"},
		{"PRICE_LEVEL_EXPENSIVE", "$$$"},
		{"PRICE_LEVEL_VERY_EXPENSIVE", "$$$$"},
	}

	for _, lvl := range levels {
		t.Run(lvl.api, func(t *testing.T) {
			p := fullPlaceFixture("p", 37.7749, -122.4194, lvl.api, nil)
			cards, _ := mapPlacesToCards([]place{p}, 37.7749, -122.4194, 500)
			if len(cards) != 1 {
				t.Fatalf("expected 1 card, got %d", len(cards))
			}
			if cards[0].PriceLevel != lvl.want {
				t.Errorf("PriceLevel = %q, want %q", cards[0].PriceLevel, lvl.want)
			}
		})
	}
}

func TestMapPlacesToCards_MultiplePhotos_UsesFirst(t *testing.T) {
	photos := []photo{
		{Name: "photos/first"},
		{Name: "photos/second"},
	}
	p := fullPlaceFixture("multi", 37.7749, -122.4194, "PRICE_LEVEL_MODERATE", photos)

	cards, _ := mapPlacesToCards([]place{p}, 37.7749, -122.4194, 500)
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].PhotoRef != "photos/first" {
		t.Errorf("PhotoRef = %q, want %q (first photo)", cards[0].PhotoRef, "photos/first")
	}
}

func TestMapPlacesToCards_DistanceEqualsRadius(t *testing.T) {
	// Place the card just inside the radius — computed such that the haversine
	// distance is slightly less than the given radius.  1 degree of latitude
	// is approximately 111,320 m, so moving north by radius / 111320 gives a
	// point that is within the radius (the latitude arc is the shortest path).
	const radius = 500.0
	boundaryLat := 37.7749 + (radius / 111320.0) // 1 deg lat ≈ 111,320 m

	p := fullPlaceFixture("boundary", boundaryLat, -122.4194, "PRICE_LEVEL_MODERATE", nil)
	cards, hitEdge := mapPlacesToCards([]place{p}, 37.7749, -122.4194, radius)

	// Because we moved due north by roughly radius metres, the haversine
	// distance should be ≤ radius (within floating-point tolerance).
	actualDist := haversineDistance(37.7749, -122.4194, boundaryLat, -122.4194)

	// Allow a tiny epsilon so the test isn't brittle across platforms.
	const epsilon = 0.5 // metres

	if actualDist > radius+epsilon {
		// The point ended up further than expected — hitEdge should be true.
		if !hitEdge {
			t.Errorf("distance %.1f > radius %.1f but hitEdge is false", actualDist, radius)
		}
		if len(cards) != 0 {
			t.Errorf("expected 0 cards when distance > radius, got %d", len(cards))
		}
	} else {
		if hitEdge {
			t.Errorf("distance %.1f <= radius %.1f but hitEdge is true", actualDist, radius)
		}
		if len(cards) != 1 {
			t.Fatalf("expected 1 card, got %d", len(cards))
		}
	}
}
