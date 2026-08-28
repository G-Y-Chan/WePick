package room

import "context"

// Service is the sole entry point transport adapters (HTTP, WS) are allowed to depend on.
// No transport package may import Repository, PlacesProvider, or Broadcaster directly.
type Service interface {
	CreateRoom(ctx context.Context) (Code, error)
	JoinRoom(ctx context.Context, code Code) (joined bool, err error)
	StartRoom(ctx context.Context, code Code, filters SearchFilters) error
	Cards(ctx context.Context, code Code) ([]Card, error)
	PhotoURL(ctx context.Context, photoRef string) (string, error)
	SubmitVote(ctx context.Context, vote Vote) error
	Connect(ctx context.Context, code Code, client Client) error
	Disconnect(ctx context.Context, code Code, client Client)
}

// ServiceOption is a functional option for the Service constructor.
type ServiceOption func(*serviceOptions)

type serviceOptions struct {
	// reserved for future use (e.g. configurable retry, timeouts)
}

// NewService is the single constructor/composition point for the domain service.
func NewService(repo Repository, places PlacesProvider, broadcaster Broadcaster, opts ...ServiceOption) Service {
	o := &serviceOptions{}
	for _, opt := range opts {
		opt(o)
	}

	return &roomService{
		repo:           repo,
		places:         places,
		broadcaster:    broadcaster,
		max:            roomCodeSpace,
		reconnectDelay: eventReconnectDelay,
	}
}
