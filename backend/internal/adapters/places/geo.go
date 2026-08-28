package places

import "math"

// =============================================================================
// Geometry Helpers — bit-for-bit identical to the original implementations in
// backend/services/places_client.go
// =============================================================================

// calculateBoundingBox returns a rectangular viewport centred on (lat, lng)
// that encloses a circle of the given radius in metres.  The equatorial Earth
// radius (6,378,137 m) is used for the angular-conversion math, matching the
// pre-refactor behaviour exactly.
func calculateBoundingBox(lat, lng float64, radiusInMeters float64) viewport {
	const earthRadius = 6378137.0 // Earth's equatorial radius in meters

	// Angular offsets in radians
	dLat := radiusInMeters / earthRadius
	dLng := radiusInMeters / (earthRadius * math.Cos(lat*math.Pi/180.0))

	return viewport{
		Low: latLng{
			Latitude:  lat - (dLat * 180.0 / math.Pi),
			Longitude: lng - (dLng * 180.0 / math.Pi),
		},
		High: latLng{
			Latitude:  lat + (dLat * 180.0 / math.Pi),
			Longitude: lng + (dLng * 180.0 / math.Pi),
		},
	}
}

// haversineDistance computes the great-circle distance in metres between two
// lat/lng points using the Haversine formula with the mean Earth radius
// (6,371,000 m).  Implementation carried over bit-for-bit from the original.
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
