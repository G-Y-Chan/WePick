package room

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"backend/internal/apperr"
)

const (
	// roomCodeSpace is the size of the 6-digit room-code space. It matches the
	// value passed to the legacy NewRoomManager (1,000,000) and bounds the
	// collision-retry loop in CreateRoom.
	roomCodeSpace = 1_000_000

	// roomTTL is the Redis key lifetime for a room. It is carried over
	// unchanged from the legacy roomTTL constant (24 hours).
	roomTTL = 24 * time.Hour

	// eventReconnectDelay is the backoff applied when the cross-process event
	// subscription cannot be established or drops unexpectedly.
	eventReconnectDelay = 2 * time.Second
)

// roomService is the concrete application service. It depends exclusively on
// the three domain ports (Repository, PlacesProvider, Broadcaster) and never
// imports an adapter or transport package.
type roomService struct {
	repo        Repository
	places      PlacesProvider
	broadcaster Broadcaster

	max            int
	reconnectDelay time.Duration
	logger         *slog.Logger
}

// EventLoopRunner is an optional capability implemented by *roomService. The
// composition root may assert this interface on a Service to start the
// cross-process pub/sub event listener.
type EventLoopRunner interface {
	StartEventListener(ctx context.Context)
}

// CreateRoom generates a 6-digit code and persists a new room, retrying on
// code collisions exactly like the legacy RoomManager.AddRoom.
func (s *roomService) CreateRoom(ctx context.Context) (Code, error) {
	for {
		code := s.generateRoomCode()

		created, err := s.repo.CreateRoom(ctx, code, roomTTL)
		if err != nil {
			return "", apperr.Wrap(apperr.CodeInternal, "failed to create room", err)
		}
		if created {
			return code, nil
		}
		// Collision: generate a new code and try again.
	}
}

// JoinRoom reports whether a room may still be joined. A started room returns
// ErrRoomAlreadyStarted (HTTP 409) instead of the legacy bare false/nil pair.
func (s *roomService) JoinRoom(ctx context.Context, code Code) (bool, error) {
	started, err := s.repo.IsStarted(ctx, code)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			return false, ErrRoomNotFound
		}
		return false, apperr.Wrap(apperr.CodeInternal, "failed to check room state", err)
	}

	if started {
		return false, ErrRoomAlreadyStarted
	}

	return true, nil
}

// StartRoom validates the search filters, fetches cards from the places
// provider, persists cards and the next-page token, then atomically marks the
// room started (which also publishes the room_started event).
func (s *roomService) StartRoom(ctx context.Context, code Code, filters SearchFilters) error {
	if err := filters.Validate(); err != nil {
		return ErrInvalidFilters
	}

	cards, nextPageToken, err := s.places.Search(ctx, filters, "")
	if err != nil {
		return apperr.Wrap(apperr.CodeUpstream, "failed to fetch places", err)
	}

	if len(cards) == 0 {
		return ErrNoPlacesFound
	}

	if err := s.repo.SaveCards(ctx, code, cards); err != nil {
		return apperr.Wrap(apperr.CodeInternal, "failed to save room cards", err)
	}

	if err := s.repo.SavePageToken(ctx, code, nextPageToken); err != nil {
		return apperr.Wrap(apperr.CodeInternal, "failed to save page token", err)
	}

	if err := s.repo.MarkStarted(ctx, code); err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			return ErrRoomNotFound
		}
		return apperr.Wrap(apperr.CodeInternal, "failed to start room", err)
	}

	return nil
}

// Cards returns the persisted venue cards for a room.
func (s *roomService) Cards(ctx context.Context, code Code) ([]Card, error) {
	cards, err := s.repo.LoadCards(ctx, code)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "failed to load room cards", err)
	}
	return cards, nil
}

// PhotoURL resolves a Google photo reference into a CDN URL for the client.
func (s *roomService) PhotoURL(ctx context.Context, photoRef string) (string, error) {
	if photoRef == "" {
		return "", apperr.New(apperr.CodeInvalid, "missing photo reference")
	}

	photoURL, err := s.places.PhotoURL(ctx, photoRef)
	if err != nil {
		return "", apperr.Wrap(apperr.CodeUpstream, "failed to fetch image", err)
	}
	return photoURL, nil
}

// SubmitVote applies the ACCEPT-only voting rule from the legacy RoomManager:
// any result other than ACCEPT is a silent no-op (logged here for visibility).
// Majority detection is delegated entirely to Repository.RegisterAcceptVote.
func (s *roomService) SubmitVote(ctx context.Context, vote Vote) error {
	if vote.Room == "" || vote.CardID == "" {
		return ErrInvalidVote
	}

	if vote.Result != VoteAccept {
		s.logger.Info("ignoring non-ACCEPT vote",
			"room_code", vote.Room.String(),
			"card_id", vote.CardID,
			"result", string(vote.Result),
		)
		return nil
	}

	majorityFound, err := s.repo.RegisterAcceptVote(ctx, vote.Room, vote.CardID)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternal, "failed to register vote", err)
	}

	if majorityFound {
		evt := Event{
			Type:   EventMajorityFound,
			Room:   vote.Room,
			CardID: vote.CardID,
		}
		if err := s.repo.PublishEvent(ctx, evt); err != nil {
			// do not fail the request if the cross-process broadcast cannot
			// be delivered.
			s.logger.Warn("failed to publish majority event",
				"room_code", vote.Room.String(),
				"card_id", vote.CardID,
				"err", err,
			)
		}
	}

	return nil
}

// Connect verifies room existence, registers the realtime client, and then
// increments the connected-client counter. If the counter update fails, the
// client is unregistered again to keep in-memory and Redis state consistent.
func (s *roomService) Connect(ctx context.Context, code Code, client Client) error {
	exists, err := s.repo.Exists(ctx, code)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternal, "failed to check room existence", err)
	}
	if !exists {
		return ErrRoomNotFound
	}

	if err := s.broadcaster.Register(code, client); err != nil {
		return apperr.Wrap(apperr.CodeInternal, "failed to register client", err)
	}

	if _, err := s.repo.IncrClientCount(ctx, code); err != nil {
		s.broadcaster.Unregister(code, client)
		return apperr.Wrap(apperr.CodeInternal, "failed to increment client count", err)
	}

	return nil
}

// Disconnect removes the realtime client and then decrements the
// connected-client counter. Decrement failures are logged, not propagated:
// the method has no return value and must not panic.
func (s *roomService) Disconnect(ctx context.Context, code Code, client Client) {
	s.broadcaster.Unregister(code, client)

	if _, err := s.repo.DecrClientCount(ctx, code); err != nil {
		s.logger.Warn("failed to decrement client count",
			"room_code", code.String(),
			"err", err,
		)
	}
}

// StartEventListener subscribes to cross-process room events and fans them out
// to connected clients. It blocks until ctx is canceled. It is intended to run
// in a dedicated goroutine owned by the composition root.
func (s *roomService) StartEventListener(ctx context.Context) {
	s.logger.Info("starting room event listener")
	defer s.logger.Info("room event listener stopped")

	for {
		sub, err := s.repo.SubscribeEvents(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			s.logger.Warn("failed to subscribe to room events; retrying",
				"retry_delay", s.reconnectDelay,
				"err", err,
			)
			if !s.sleepWithContext(ctx) {
				return
			}
			continue
		}

		s.logger.Info("subscribed to room events")
		s.consumeEvents(ctx, sub)
		_ = sub.Close()

		if ctx.Err() != nil {
			return
		}

		s.logger.Warn("room event subscription closed unexpectedly; reconnecting",
			"retry_delay", s.reconnectDelay,
		)
		if !s.sleepWithContext(ctx) {
			return
		}
	}
}

// consumeEvents drains a single subscription until the context is canceled or
// the subscription's event channel is closed.
func (s *roomService) consumeEvents(ctx context.Context, sub EventSubscription) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-sub.Events():
			if !ok {
				return
			}
			s.handleEvent(evt)
		}
	}
}

// handleEvent translates a cross-process domain event into a realtime fan-out.
func (s *roomService) handleEvent(evt Event) {
	switch evt.Type {
	case EventRoomStarted:
		s.logger.Info("received room_started event", "room_code", evt.Room.String())
		s.broadcaster.BroadcastRoomStarted(evt.Room)
	case EventMajorityFound:
		s.logger.Info("received majority_found event",
			"room_code", evt.Room.String(),
			"card_id", evt.CardID,
		)
		s.broadcaster.BroadcastMajorityFound(evt.Room, evt.CardID)
	default:
		s.logger.Warn("ignoring unknown room event", "event_type", string(evt.Type))
	}
}

// sleepWithContext waits for reconnectDelay or until ctx is canceled.
func (s *roomService) sleepWithContext(ctx context.Context) bool {
	timer := time.NewTimer(s.reconnectDelay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// generateRoomCode produces a zero-padded 6-digit code using the same
// math/rand approach as the legacy GenerateRoomCode.
func (s *roomService) generateRoomCode() Code {
	max := s.max
	if max <= 0 {
		max = roomCodeSpace
	}
	return Code(fmt.Sprintf("%06d", rand.Intn(max)))
}
