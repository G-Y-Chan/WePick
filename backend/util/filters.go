package util

type Filters struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Radius    int     `json:"radius"`
	MaxPrice  int     `json:"maxPrice"`
	Category  string  `json:"category"`
	OpenNow   bool    `json:"openNow"`
}
