package room

import "errors"

type Filters struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Radius    int     `json:"radius"`
	MaxPrice  int     `json:"maxPrice"`
	Category  string  `json:"category"`
	OpenNow   bool    `json:"openNow"`
}

func (f Filters) Validate() error {
	if f.Latitude < -90 || f.Latitude > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	if f.Longitude < -180 || f.Longitude > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	if f.Radius <= 0 {
		return errors.New("radius must be greater than 0")
	}
	if f.Radius > 50000 {
		return errors.New("radius must not exceed 50000 meters")
	}
	switch f.Category {
	case "restaurant", "cafe", "bar", "":
		// valid
	default:
		return errors.New("category must be one of: restaurant, cafe, bar")
	}
	return nil
}