package http

import (
	"backend/internal/domain/room"
)

// =============================================================================
// Wire-compatible DTOs
// =============================================================================
// Field names and JSON tags below are copied exactly from the legacy util
// package so the public HTTP/WS contract remains byte-for-byte compatible.
// Do not "clean up" the casing or rename fields without a separately-flagged
// breaking-change decision.
// =============================================================================

// filtersDTO mirrors util.Filters exactly.
type filtersDTO struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Radius    int     `json:"radius"`
	MaxPrice  int     `json:"maxPrice"`
	Category  string  `json:"category"`
	OpenNow   bool    `json:"openNow"`
}

func (d filtersDTO) toDomain() (room.SearchFilters, error) {
	f := room.SearchFilters{
		Latitude:  d.Latitude,
		Longitude: d.Longitude,
		RadiusM:   d.Radius,
		MaxPrice:  d.MaxPrice,
		Category:  room.Category(d.Category),
		OpenNow:   d.OpenNow,
	}
	if err := f.Validate(); err != nil {
		return room.SearchFilters{}, err
	}
	return f, nil
}

// cardDTO mirrors util.Card exactly (same field names, same tags).
type cardDTO struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Category    string  `json:"category"`
	PriceLevel  string  `json:"priceLevel"`
	Rating      float64 `json:"rating"`
	ReviewCount int     `json:"reviewCount"`
	OpenNow     bool    `json:"openNow"`
	Summary     string  `json:"summary"`
	Address     string  `json:"address"`
	PhotoName   string  `json:"photoName,omitempty"`
}

func cardDTOFromDomain(c room.Card) cardDTO {
	return cardDTO{
		ID:          c.ID,
		Title:       c.Title,
		Category:    c.Category,
		PriceLevel:  c.PriceLevel,
		Rating:      c.Rating,
		ReviewCount: c.ReviewCount,
		OpenNow:     c.OpenNow,
		Summary:     c.Summary,
		Address:     c.Address,
		PhotoName:   c.PhotoRef,
	}
}

func cardDTOsFromDomain(cs []room.Card) []cardDTO {
	if cs == nil {
		return nil
	}
	out := make([]cardDTO, len(cs))
	for i, c := range cs {
		out[i] = cardDTOFromDomain(c)
	}
	return out
}

// messageDTO mirrors util.Message exactly, INCLUDING the capitalized JSON tags
// (existing clients depend on this casing).
type messageDTO struct {
	Header  string    `json:"Header"`
	Body    *string   `json:"Body,omitempty"`
	VoteObj *voteDTO  `json:"VoteObj,omitempty"`
	Cards   []cardDTO `json:"Cards,omitempty"`
}

// voteDTO mirrors util.Vote exactly. The original struct has no json tags at
// all, so the wire keys are the bare Go field names "Id"/"Result"/"Room".
type voteDTO struct {
	Id     string
	Result string
	Room   string
}

// errorResponseDTO mirrors util.ErrorResponse exactly, including the legacy
// "Message" field (kept for wire compatibility in this phase; flagged for
// removal as a separate v2 cleanup).
type errorResponseDTO struct {
	Header  string `json:"Header"`
	Body    string `json:"Body"`
	Message string `json:"Message"`
}

// startRoomPayload mirrors the legacy handlers.StartRoomPayload shape.
type startRoomPayload struct {
	Filters filtersDTO `json:"filters"`
}
