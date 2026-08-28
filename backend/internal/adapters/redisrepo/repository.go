package redisrepo

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"backend/internal/apperr"
	"backend/internal/domain/room"
)

// Compile-time contract check: Repository satisfies room.Repository.
var _ room.Repository = (*Repository)(nil)

// ---------------------------------------------------------------------------
// Repository implements room.Repository backed by Redis.
// All method signatures mirror room.Repository exactly.
// Redis key layout, hash field names, Lua script bodies, and TTL are carried
// over unchanged from the legacy room.RoomRepository.
// ---------------------------------------------------------------------------

type Repository struct {
	rdb *redis.Client
}

// New creates a Repository backed by the given Redis client.
func New(rdb *redis.Client) *Repository {
	return &Repository{rdb: rdb}
}

// ---------------------------------------------------------------------------
// Room lifecycle
// ---------------------------------------------------------------------------

// CreateRoom atomically creates a new room in Redis if it does not already
// exist. Returns created=false when the code collides with an existing room.
func (r *Repository) CreateRoom(ctx context.Context, code room.Code, ttl time.Duration) (bool, error) {
	res, err := createRoomScript.Run(
		ctx,
		r.rdb,
		[]string{roomKey(code)},
		int(ttl.Seconds()),
	).Int()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

// Exists reports whether a room with the given code exists in Redis.
func (r *Repository) Exists(ctx context.Context, code room.Code) (bool, error) {
	res, err := r.rdb.Exists(ctx, roomKey(code)).Result()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

// IsStarted reports whether the room has been started.
// Returns a domain sentinel error when the room does not exist.
func (r *Repository) IsStarted(ctx context.Context, code room.Code) (bool, error) {
	startedStr, err := r.rdb.HGet(ctx, roomKey(code), "started").Result()
	if err == redis.Nil {
		return false, room.ErrRoomNotFound
	}
	if err != nil {
		return false, err
	}
	started, err := strconv.ParseBool(startedStr)
	if err != nil {
		return false, err
	}
	return started, nil
}

// MarkStarted atomically marks the room as started and publishes a
// room_started event via Redis Pub/Sub. The Lua script preserves the
// atomicity contract from the legacy codebase.
func (r *Repository) MarkStarted(ctx context.Context, code room.Code) error {
	event := room.Event{
		Type: room.EventRoomStarted,
		Room: code,
	}
	payload, err := json.Marshal(redisEventFromDomain(event))
	if err != nil {
		return err
	}

	res, err := startRoomScript.Run(
		ctx,
		r.rdb,
		[]string{roomKey(code)},
		roomEventsChannel,
		payload,
	).Int()
	if err != nil {
		return err
	}
	if res == 0 {
		return room.ErrRoomNotFound
	}
	return nil
}

// ---------------------------------------------------------------------------
// Cards
// ---------------------------------------------------------------------------

// SaveCards stores the room's venue cards as a JSON blob in the room hash.
func (r *Repository) SaveCards(ctx context.Context, code room.Code, cards []room.Card) error {
	// Marshal through the wire-compatible redisCard shape so the stored JSON
	// matches the legacy format exactly (photoName key, not PhotoRef).
	rcs := make([]redisCard, len(cards))
	for i, c := range cards {
		rcs[i] = redisCardFromDomain(c)
	}
	data, err := json.Marshal(rcs)
	if err != nil {
		return err
	}
	return r.rdb.HSet(ctx, roomKey(code), "cards", data).Err()
}

// LoadCards retrieves the room's venue cards from the room hash.
func (r *Repository) LoadCards(ctx context.Context, code room.Code) ([]room.Card, error) {
	data, err := r.rdb.HGet(ctx, roomKey(code), "cards").Bytes()
	if err == redis.Nil {
		return nil, nil // No cards stored yet
	}
	if err != nil {
		return nil, err
	}

	var rcs []redisCard
	if err := json.Unmarshal(data, &rcs); err != nil {
		return nil, err
	}

	cards := make([]room.Card, len(rcs))
	for i, rc := range rcs {
		cards[i] = rc.toDomain()
	}
	return cards, nil
}

// ---------------------------------------------------------------------------
// Page token
// ---------------------------------------------------------------------------

// SavePageToken stores the Google Places pagination token.
func (r *Repository) SavePageToken(ctx context.Context, code room.Code, token string) error {
	return r.rdb.HSet(ctx, roomKey(code), "pageToken", token).Err()
}

// LoadPageToken retrieves the Google Places pagination token.
func (r *Repository) LoadPageToken(ctx context.Context, code room.Code) (string, error) {
	token, err := r.rdb.HGet(ctx, roomKey(code), "pageToken").Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return token, nil
}

// ---------------------------------------------------------------------------
// Client count
// ---------------------------------------------------------------------------

// IncrClientCount atomically increments the connected-client counter.
func (r *Repository) IncrClientCount(ctx context.Context, code room.Code) (int64, error) {
	count, err := r.rdb.HIncrBy(ctx, roomKey(code), "client_count", 1).Result()
	if err != nil {
		return 0, err
	}
	return count, nil
}

// DecrClientCount atomically decrements the connected-client counter.
func (r *Repository) DecrClientCount(ctx context.Context, code room.Code) (int64, error) {
	count, err := r.rdb.HIncrBy(ctx, roomKey(code), "client_count", -1).Result()
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ClientCount returns the current connected-client count.
func (r *Repository) ClientCount(ctx context.Context, code room.Code) (int64, error) {
	countStr, err := r.rdb.HGet(ctx, roomKey(code), "client_count").Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	count, err := strconv.ParseInt(countStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ---------------------------------------------------------------------------
// Voting
// ---------------------------------------------------------------------------

// RegisterAcceptVote atomically increments the accept-count for cardID and
// reports whether every currently-connected client has now accepted (majority).
// Backed by the same incrementAndCheckScript Lua script as the legacy code.
func (r *Repository) RegisterAcceptVote(ctx context.Context, code room.Code, cardID string) (bool, error) {
	result, err := incrementAndCheckScript.Run(
		ctx,
		r.rdb,
		[]string{voteKey(code), roomKey(code)},
		cardID,
		int64(RoomTTL),
	).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

// ---------------------------------------------------------------------------
// Pub / Sub events
// ---------------------------------------------------------------------------

// PublishEvent publishes a domain Event on the Redis Pub/Sub channel.
// The payload matches the legacy wire format so that old and new subscribers
// (potentially running side by side during migration) see identical payloads.
func (r *Repository) PublishEvent(ctx context.Context, event room.Event) error {
	re := redisEventFromDomain(event)
	payload, err := json.Marshal(re)
	if err != nil {
		return err
	}
	return r.rdb.Publish(ctx, roomEventsChannel, payload).Err()
}

// SubscribeEvents creates an EventSubscription that translates raw Redis
// Pub/Sub messages into domain room.Event values.
func (r *Repository) SubscribeEvents(ctx context.Context) (room.EventSubscription, error) {
	pubsub := r.rdb.Subscribe(ctx, roomEventsChannel)

	// Ensure the subscription is ready before returning.
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return nil, err
	}

	events := make(chan room.Event, 64)
	childCtx, cancel := context.WithCancel(ctx)

	go func() {
		defer close(events)
		defer pubsub.Close()
		defer cancel()

		ch := pubsub.Channel()
		for {
			select {
			case <-childCtx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var re redisEvent
				if err := json.Unmarshal([]byte(msg.Payload), &re); err != nil {
					// Malformed payload — skip, don't crash the subscription.
					continue
				}
				evt := re.toDomain()
				select {
				case events <- evt:
				case <-childCtx.Done():
					return
				}
			}
		}
	}()

	return &eventSubscription{
		events: events,
		cancel: cancel,
		pubsub: pubsub,
	}, nil
}

// ---------------------------------------------------------------------------
// EventSubscription
// ---------------------------------------------------------------------------

type eventSubscription struct {
	events chan room.Event
	cancel context.CancelFunc
	pubsub *redis.PubSub
}

func (es *eventSubscription) Events() <-chan room.Event { return es.events }

func (es *eventSubscription) Close() error {
	es.cancel()
	_ = es.pubsub.Close()
	return nil
}

// ---------------------------------------------------------------------------
// Wire-format DTOs (package-private)
// ---------------------------------------------------------------------------

// redisCard mirrors the legacy util.Card JSON shape exactly, preserving the
// "photoName" key so that stored Redis data remains readable by old code.
type redisCard struct {
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

func redisCardFromDomain(c room.Card) redisCard {
	return redisCard{
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

func (rc redisCard) toDomain() room.Card {
	return room.Card{
		ID:          rc.ID,
		Title:       rc.Title,
		Category:    rc.Category,
		PriceLevel:  rc.PriceLevel,
		Rating:      rc.Rating,
		ReviewCount: rc.ReviewCount,
		OpenNow:     rc.OpenNow,
		Summary:     rc.Summary,
		Address:     rc.Address,
		PhotoRef:    rc.PhotoName,
	}
}

// redisEvent mirrors the legacy room.RoomEvent JSON shape exactly, preserving
// the "voteId" key so that the wire format is compatible with old code.
type redisEvent struct {
	Type   string `json:"type"`
	Room   string `json:"room"`
	VoteID string `json:"voteId,omitempty"`
}

func redisEventFromDomain(e room.Event) redisEvent {
	return redisEvent{
		Type:   string(e.Type),
		Room:   e.Room.String(),
		VoteID: e.CardID,
	}
}

func (re redisEvent) toDomain() room.Event {
	return room.Event{
		Type:   room.EventType(re.Type),
		Room:   room.Code(re.Room),
		CardID: re.VoteID,
	}
}

// Ensure apperr is referenced (used by IsStarted returning sentinel errors).
var _ = apperr.CodeInvalid // compile-time usage guard
