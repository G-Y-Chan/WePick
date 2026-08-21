package room

import (
	"context"
	"errors"
)

var (
	ErrRoomNotFound        = errors.New("room not found")
	ErrRoomAlreadyStarted  = errors.New("room already started")
	ErrVoteRoomRequired    = errors.New("room is required")
	ErrVoteIDRequired      = errors.New("vote id is required")
	ErrNoPlacesFound       = errors.New("no places found within the specified area")
)

// Repository defines the interface for Redis-backed room persistence.
type Repository interface {
	CreateRoom(ctx context.Context, code string) (bool, error)
	CheckIfRoomExists(ctx context.Context, code string) (bool, error)
	CheckIfRoomStarted(ctx context.Context, code string) (bool, error)
	StartRoom(ctx context.Context, code string) error
	IncrementAcceptVote(ctx context.Context, roomID, voteID, clientID string) (bool, error)
	IncrementRoomClientCount(ctx context.Context, roomCode string) (int64, error)
	DecrementRoomClientCount(ctx context.Context, roomCode string) (int64, error)
	GetRoomClientCount(ctx context.Context, roomCode string) (int64, error)
	PublishMajorityFound(ctx context.Context, roomID, voteID string) error
	SubscribeToRoomEvents(ctx context.Context) (EventSubscription, error)
	SetRoomCards(ctx context.Context, code string, cards []Card) error
	GetRoomCards(ctx context.Context, code string) ([]Card, error)
	SetPageToken(ctx context.Context, code string, token string) error
}

// EventSubscription wraps a Redis pubsub and delivers decoded RoomEvents.
type EventSubscription interface {
	Events() <-chan RoomEvent
	Close() error
}

// PlacesAPI defines the interface for the Google Places client.
type PlacesAPI interface {
	FetchCards(filters Filters, pageToken string) ([]Card, string, error)
	GetPhotoURL(photoName string) (string, error)
}

// Card is a domain-level place card returned from the Places API.
type Card struct {
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

// Vote represents a vote action from a client.
type Vote struct {
	Id       string
	Result   string
	Room     string
	ClientID string
}