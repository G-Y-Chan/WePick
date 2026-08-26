package room

import (
	"context"
	"time"
)

// Repository is the persistence port. Implemented by adapters/redisrepo.Repository.
type Repository interface {
	CreateRoom(ctx context.Context, code Code, ttl time.Duration) (created bool, err error)
	Exists(ctx context.Context, code Code) (bool, error)
	IsStarted(ctx context.Context, code Code) (bool, error)
	MarkStarted(ctx context.Context, code Code) error

	SaveCards(ctx context.Context, code Code, cards []Card) error
	LoadCards(ctx context.Context, code Code) ([]Card, error)

	SavePageToken(ctx context.Context, code Code, token string) error
	LoadPageToken(ctx context.Context, code Code) (string, error)

	IncrClientCount(ctx context.Context, code Code) (int64, error)
	DecrClientCount(ctx context.Context, code Code) (int64, error)
	ClientCount(ctx context.Context, code Code) (int64, error)

	// RegisterAcceptVote atomically increments the accept-count for cardID in code
	// and reports whether every currently-connected client has now accepted (majority).
	// Backed by the existing Lua script for atomicity — logic must not change.
	RegisterAcceptVote(ctx context.Context, code Code, cardID string) (majorityReached bool, err error)

	PublishEvent(ctx context.Context, event Event) error
	SubscribeEvents(ctx context.Context) (EventSubscription, error)
}

// EventSubscription delivers Events from the pub/sub bus to the caller.
type EventSubscription interface {
	Events() <-chan Event
	Close() error
}

// PlacesProvider is the external venue-search port. Implemented by adapters/places.Client.
type PlacesProvider interface {
	Search(ctx context.Context, filters SearchFilters, pageToken string) (cards []Card, nextPageToken string, err error)
	PhotoURL(ctx context.Context, photoRef string) (string, error)
}

// Broadcaster is the realtime fan-out port. Implemented by adapters/wshub.Hub.
// The domain/service layer knows nothing about WebSockets through this interface.
type Broadcaster interface {
	Register(code Code, client Client) error
	Unregister(code Code, client Client)
	BroadcastRoomStarted(code Code)
	BroadcastMajorityFound(code Code, cardID string)
}

// Client is the minimal contract a transport connection must satisfy to receive pushes.
// Implemented by adapters/wshub.Conn.
type Client interface {
	ID() string
	Send(evt OutboundEvent) error // MUST be non-blocking
}
