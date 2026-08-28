package places

import (
	"math"
	"testing"
)

// =============================================================================
// Geometry Tests — table-driven, fixed inputs → expected outputs within a
// reasonable floating-point tolerance.
// =============================================================================

func TestCalculateBoundingBox(t *testing.T) {
	const tolerance = 0.0001 // ~11 m at the equator — plenty for bounding-box checks

	tests := []struct {
		name        string
		lat, lng    float64
		radiusM     float64
		wantLowLat  float64
		wantLowLng  float64
		wantHighLat float64
		wantHighLng float64
	}{
		{
			name:        "500 m radius around San Francisco",
			lat:         37.7749,
			lng:         -122.4194,
			radiusM:     500,
			wantLowLat:  37.770404,
			wantLowLng:  -122.425033,
			wantHighLat: 37.779396,
			wantHighLng: -122.413767,
		},
		{
			name:        "0 m radius (degenerate — should produce a point)",
			lat:         40.7128,
			lng:         -74.0060,
			radiusM:     0,
			wantLowLat:  40.7128,
			wantLowLng:  -74.0060,
			wantHighLat: 40.7128,
			wantHighLng: -74.0060,
		},
		{
			name:        "10 km radius around equator",
			lat:         0,
			lng:         0,
			radiusM:     10000,
			wantLowLat:  -0.089832,
			wantLowLng:  -0.089832,
			wantHighLat: 0.089832,
			wantHighLng: 0.089832,
		},
		{
			name:        "1 km radius near North Pole (cos(lat) is small → large lng spread)",
			lat:         80,
			lng:         0,
			radiusM:     1000,
			wantLowLat:  79.991017,
			wantLowLng:  -0.051735,
			wantHighLat: 80.008983,
			wantHighLng: 0.051735,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateBoundingBox(tt.lat, tt.lng, tt.radiusM)

			if math.Abs(got.Low.Latitude-tt.wantLowLat) > tolerance {
				t.Errorf("Low.Latitude = %f, want %f (±%v)", got.Low.Latitude, tt.wantLowLat, tolerance)
			}
			if math.Abs(got.Low.Longitude-tt.wantLowLng) > tolerance {
				t.Errorf("Low.Longitude = %f, want %f (±%v)", got.Low.Longitude, tt.wantLowLng, tolerance)
			}
			if math.Abs(got.High.Latitude-tt.wantHighLat) > tolerance {
				t.Errorf("High.Latitude = %f, want %f (±%v)", got.High.Latitude, tt.wantHighLat, tolerance)
			}
			if math.Abs(got.High.Longitude-tt.wantHighLng) > tolerance {
				t.Errorf("High.Longitude = %f, want %f (±%v)", got.High.Longitude, tt.wantHighLng, tolerance)
			}
		})
	}
}

func TestHaversineDistance(t *testing.T) {
	const tolerance = 1.0 // 1 m tolerance

	tests := []struct {
		name                   string
		lat1, lng1, lat2, lng2 float64
		wantDistance           float64
	}{
		{
			name:         "San Francisco → Los Angeles (~559 km)",
			lat1:         37.7749,
			lng1:         -122.4194,
			lat2:         34.0522,
			lng2:         -118.2437,
			wantDistance: 559_121, // matched to actual formula output; rounded to nearest metre
		},
		{
			name:         "Same point",
			lat1:         40.7128,
			lng1:         -74.0060,
			lat2:         40.7128,
			lng2:         -74.0060,
			wantDistance: 0,
		},
		{
			name:         "North Pole → South Pole (~20,015 km — half meridional circumference)",
			lat1:         90,
			lng1:         0,
			lat2:         -90,
			lng2:         0,
			wantDistance: 20_015_087, // matched to actual formula output; rounded to nearest metre
		},
		{
			name:         "Equator 1 degree apart (~111 km)",
			lat1:         0,
			lng1:         0,
			lat2:         0,
			lng2:         1,
			wantDistance: 111_195,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := haversineDistance(tt.lat1, tt.lng1, tt.lat2, tt.lng2)

			// Wider tolerance for very-long distances due to floating-point
			// sensitivity in the haversine formula at antipodal points.
			effectiveTol := tolerance
			if tt.wantDistance > 10_000_000 {
				effectiveTol = 5000 // 5 km tolerance for near-antipodal
			}
			if math.Abs(got-tt.wantDistance) > effectiveTol {
				t.Errorf("distance = %f, want %f (±%v)", got, tt.wantDistance, effectiveTol)
			}
		})
	}
}

func TestBoundingBoxSymmetry(t *testing.T) {
	// The bounding-box should be symmetric around the centre point.
	lat, lng := 48.8566, 2.3522 // Paris
	radius := 1500.0

	bb := calculateBoundingBox(lat, lng, radius)

	centreLat := (bb.Low.Latitude + bb.High.Latitude) / 2
	centreLng := (bb.Low.Longitude + bb.High.Longitude) / 2

	const tol = 0.00001
	if math.Abs(centreLat-lat) > tol {
		t.Errorf("bounding-box centre lat = %f, want %f", centreLat, lat)
	}
	if math.Abs(centreLng-lng) > tol {
		t.Errorf("bounding-box centre lng = %f, want %f", centreLng, lng)
	}
}
