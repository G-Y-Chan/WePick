package room

import (
	"backend/internal/apperr"
	"fmt"
	"strconv"
	"time"
)

// Code is a validated 6-digit room code (matches the existing "%06d" generation format).
type Code string

// ParseCode validates a raw path/query string into a Code.
// Returns an apperr.Error with CodeInvalid on malformed input.
func ParseCode(raw string) (Code, error) {
	if len(raw) != 6 {
		return "", apperr.New(apperr.CodeInvalid, "room code must be exactly 6 digits")
	}
	if _, err := strconv.Atoi(raw); err != nil {
		return "", apperr.New(apperr.CodeInvalid, "room code must be numeric")
	}
	return Code(raw), nil
}

func (c Code) String() string {
	return string(c)
}

// Room is the domain aggregate representing a swipe-to-pick session.
type Room struct {
	Code        Code
	Started     bool
	ClientCount int64
	Cards       []Card
	PageToken   string
	CreatedAt   time.Time
}

// Card is a venue presented to users for voting.
type Card struct {
	ID          string
	Title       string
	Category    string
	PriceLevel  string
	Rating      float64
	ReviewCount int
	OpenNow     bool
	Summary     string
	Address     string
	PhotoRef    string // Google "photo name" resource reference, not a resolved URL
}

// Category narrows the venue search to a specific place type.
type Category string

const (
	CategoryRestaurant Category = "restaurant"
	CategoryCafe       Category = "cafe"
	CategoryBar        Category = "bar"
)

// SearchFilters carries the user's venue-search constraints.
type SearchFilters struct {
	Latitude  float64
	Longitude float64
	RadiusM   int
	MaxPrice  int
	Category  Category
	OpenNow   bool
}

// Validate enforces domain invariants on search filters.
// This is the domain-layer safety net, independent of whatever transport called it.
func (f SearchFilters) Validate() error {
	if f.Latitude < -90 || f.Latitude > 90 {
		return apperr.New(apperr.CodeInvalid, fmt.Sprintf("latitude %f out of range [-90, 90]", f.Latitude))
	}
	if f.Longitude < -180 || f.Longitude > 180 {
		return apperr.New(apperr.CodeInvalid, fmt.Sprintf("longitude %f out of range [-180, 180]", f.Longitude))
	}
	if f.RadiusM <= 0 || f.RadiusM > 50_000 {
		return apperr.New(apperr.CodeInvalid, fmt.Sprintf("radius %d out of range (0, 50000]", f.RadiusM))
	}
	if f.MaxPrice < 0 || f.MaxPrice > 4 {
		return apperr.New(apperr.CodeInvalid, fmt.Sprintf("maxPrice %d out of range [0, 4]", f.MaxPrice))
	}
	switch f.Category {
	case CategoryRestaurant, CategoryCafe, CategoryBar, "":
		// valid
	default:
		return apperr.New(apperr.CodeInvalid, fmt.Sprintf("unknown category %q", f.Category))
	}
	return nil
}

// VoteResult expresses whether a participant accepted or rejected a Card.
type VoteResult string

const (
	VoteAccept VoteResult = "ACCEPT"
	VoteReject VoteResult = "REJECT"
)

// ParseVoteResult normalizes a raw vote-result string.
func ParseVoteResult(raw string) (VoteResult, error) {
	switch VoteResult(raw) {
	case VoteAccept:
		return VoteAccept, nil
	case VoteReject:
		return VoteReject, nil
	default:
		return "", apperr.New(apperr.CodeInvalid, fmt.Sprintf("unknown vote result %q", raw))
	}
}

// Vote captures a single participant's choice on a specific Card.
type Vote struct {
	Room   Code
	CardID string
	Result VoteResult
}
