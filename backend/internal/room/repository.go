package room

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	roomEventsChannel = "room_events"
	fieldStarted      = "started"
	fieldCards        = "cards"
	fieldPageToken    = "pageToken"
	fieldClientCount  = "client_count"

	roomTTL = 24 * time.Hour
)

func roomKey(code string) string {
	return "room:" + code
}

func voteKey(roomID string) string {
	return "votes:" + roomID
}

func voterSetKey(roomID, voteID string) string {
	return fmt.Sprintf("voters:%s:%s", roomID, voteID)
}

// incrementAndCheckScript increments a vote only if the client hasn't voted yet (SADD check).
var incrementAndCheckScript = redis.NewScript(`
local voteKey = KEYS[1]
local roomKey = KEYS[2]
local voterSetKey = KEYS[3]
local voteID = ARGV[1]
local clientID = ARGV[2]
local ttl = ARGV[3]

-- Check if this client already voted for this voteID
local added = redis.call("SADD", voterSetKey, clientID)
if added == 0 then
	return -1  -- already voted
end
redis.call("EXPIRE", voterSetKey, ttl)

-- Increment the vote
local newCount = redis.call("HINCRBY", voteKey, voteID, 1)
redis.call("EXPIRE", voteKey, ttl)

-- Fetch the live client count directly from the room hash
local numClientsStr = redis.call("HGET", roomKey, "client_count")
local numClients = 0
if numClientsStr then
	numClients = tonumber(numClientsStr)
end

-- Check for unanimity (100% acceptance)
if numClients > 0 and newCount >= numClients then
	return 1
else
	return 0
end
`)

var createRoomScript = redis.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 1 then
	return 0
end
redis.call("HSET", KEYS[1], "started", "false", "client_count", 0)
redis.call("EXPIRE", KEYS[1], ARGV[1])
return 1
`)

var startRoomScript = redis.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 0 then
	return 0
end
redis.call("HSET", KEYS[1], "started", "true")
redis.call("PUBLISH", ARGV[1], ARGV[2])
return 1
`)

type RoomRepository struct {
	rdb *redis.Client
}

func NewRoomRepository(rdb *redis.Client) *RoomRepository {
	return &RoomRepository{rdb: rdb}
}

func (rr *RoomRepository) CreateRoom(ctx context.Context, code string) (bool, error) {
	res, err := createRoomScript.Run(
		ctx, rr.rdb,
		[]string{roomKey(code)},
		int(roomTTL.Seconds()),
	).Int()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

func (rr *RoomRepository) CheckIfRoomExists(ctx context.Context, code string) (bool, error) {
	res, err := rr.rdb.Exists(ctx, roomKey(code)).Result()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

func (rr *RoomRepository) CheckIfRoomStarted(ctx context.Context, code string) (bool, error) {
	startedStr, err := rr.rdb.HGet(ctx, roomKey(code), fieldStarted).Result()
	if err == redis.Nil {
		return false, ErrRoomNotFound
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

func (rr *RoomRepository) StartRoom(ctx context.Context, code string) error {
	event := RoomEvent{Type: "room_started", Room: code}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	res, err := startRoomScript.Run(
		ctx, rr.rdb,
		[]string{roomKey(code)},
		roomEventsChannel, string(payload),
	).Int()
	if err != nil {
		return err
	}
	if res == 0 {
		return ErrRoomNotFound
	}
	return nil
}

func (rr *RoomRepository) IncrementAcceptVote(ctx context.Context, roomID, voteID, clientID string) (bool, error) {
	result, err := incrementAndCheckScript.Run(
		ctx, rr.rdb,
		[]string{voteKey(roomID), roomKey(roomID), voterSetKey(roomID, voteID)},
		voteID, clientID, int(roomTTL.Seconds()),
	).Int()
	if err != nil {
		return false, err
	}
	switch result {
	case -1:
		slog.Warn("duplicate vote rejected", "roomID", roomID, "voteID", voteID, "clientID", clientID)
		return false, nil
	case 1:
		return true, nil
	default:
		return false, nil
	}
}

func (rr *RoomRepository) IncrementRoomClientCount(ctx context.Context, roomCode string) (int64, error) {
	return rr.rdb.HIncrBy(ctx, roomKey(roomCode), fieldClientCount, 1).Result()
}

func (rr *RoomRepository) DecrementRoomClientCount(ctx context.Context, roomCode string) (int64, error) {
	return rr.rdb.HIncrBy(ctx, roomKey(roomCode), fieldClientCount, -1).Result()
}

func (rr *RoomRepository) GetRoomClientCount(ctx context.Context, roomCode string) (int64, error) {
	countStr, err := rr.rdb.HGet(ctx, roomKey(roomCode), fieldClientCount).Result()
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

func (rr *RoomRepository) PublishMajorityFound(ctx context.Context, roomID, voteID string) error {
	event := RoomEvent{Type: "majority_found", Room: roomID, VoteID: voteID}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return rr.rdb.Publish(ctx, roomEventsChannel, string(payload)).Err()
}

func (rr *RoomRepository) SubscribeToRoomEvents(ctx context.Context) (EventSubscription, error) {
	pubsub := rr.rdb.Subscribe(ctx, roomEventsChannel)
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return nil, err
	}
	return &eventSubscriptionImpl{pubsub: pubsub}, nil
}

type eventSubscriptionImpl struct {
	pubsub *redis.PubSub
}

func (s *eventSubscriptionImpl) Events() <-chan RoomEvent {
	ch := make(chan RoomEvent, 64)
	go func() {
		redisCh := s.pubsub.Channel()
		for msg := range redisCh {
			var event RoomEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				slog.Warn("invalid room event", "error", err)
				continue
			}
			ch <- event
		}
		close(ch)
	}()
	return ch
}

func (s *eventSubscriptionImpl) Close() error {
	return s.pubsub.Close()
}

func (rr *RoomRepository) SetRoomCards(ctx context.Context, code string, cards []Card) error {
	data, err := json.Marshal(cards)
	if err != nil {
		return err
	}
	return rr.rdb.HSet(ctx, roomKey(code), fieldCards, data).Err()
}

func (rr *RoomRepository) GetRoomCards(ctx context.Context, code string) ([]Card, error) {
	data, err := rr.rdb.HGet(ctx, roomKey(code), fieldCards).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var cards []Card
	if err := json.Unmarshal(data, &cards); err != nil {
		return nil, err
	}
	return cards, nil
}

func (rr *RoomRepository) SetPageToken(ctx context.Context, code, token string) error {
	return rr.rdb.HSet(ctx, roomKey(code), fieldPageToken, token).Err()
}

func (rr *RoomRepository) GetPageToken(ctx context.Context, code string) (string, error) {
	token, err := rr.rdb.HGet(ctx, roomKey(code), fieldPageToken).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return token, nil
}