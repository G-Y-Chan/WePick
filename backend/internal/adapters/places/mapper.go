package places

import "backend/internal/domain/room"

// =============================================================================
// Anti-Corruption Layer — Google wire shape → domain model
// =============================================================================
// Every function below is carried over from backend/services/places_client.go
// with the sole intentional change being that the target type is room.Card
// (PhotoRef) instead of util.Card (PhotoName).  All logic branches, the
// order of iteration, the early-break-on-edge behaviour, and every mapping
// decision are preserved exactly.

// mapCategoryToTextQuery converts a domain Category into the free-form text
// query string Google Places expects.
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

// mapCategoryToIncludedType converts a domain Category into a Google Place
// Type enum value (must match Google's official Place Types exactly).
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

// mapPriceLevel translates a Google PRICE_LEVEL_* enum string into a
// human-readable display label.
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

// mapPlacesToCards converts a raw Google Places result set into domain Cards
// and reports whether the result set hit the radius boundary (hitEdge), which
// signals that pagination should stop.
//
// Behavioural parity with the original mapPlacesToCards (services/places_client.go):
//   - Iterates in order; the first place whose haversine distance exceeds the
//     radius triggers hitEdge = true and halts the loop entirely (including
//     skipping all subsequent places, even those that might be within radius).
//   - The first photo's Name is used as the photo reference; if the slice is
//     empty the reference is the empty string.
func mapPlacesToCards(places []place, centerLat, centerLng float64, maxRadius float64) ([]room.Card, bool) {
	cards := make([]room.Card, 0, len(places))
	hitEdge := false

	for _, p := range places {
		distance := haversineDistance(centerLat, centerLng, p.Location.Latitude, p.Location.Longitude)

		if distance > maxRadius {
			hitEdge = true
			break
		}

		priceDisplay := mapPriceLevel(p.PriceLevel)

		photoRef := ""
		if len(p.Photos) > 0 {
			photoRef = p.Photos[0].Name
		}

		cards = append(cards, room.Card{
			ID:          p.ID,
			Title:       p.DisplayName.Text,
			Category:    p.PrimaryTypeDisplayName.Text,
			PriceLevel:  priceDisplay,
			Rating:      p.Rating,
			ReviewCount: p.UserRatingCount,
			OpenNow:     p.CurrentOpeningHours.OpenNow,
			Summary:     p.EditorialSummary.Text,
			Address:     p.ShortFormattedAddress,
			PhotoRef:    photoRef,
		})
	}
	return cards, hitEdge
}
