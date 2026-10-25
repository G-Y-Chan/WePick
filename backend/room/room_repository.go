package room

import (
	"backend/util"
	"context"
	"encoding/json"
	"fmt"
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

// ==========================================
// Key Generators
// ==========================================

func roomKey(code string) string {
	return "room:" + code
}

func voteKey(roomID string) string {
	return fmt.Sprintf("votes:%s", roomID)
}

// ==========================================
// Lua Scripts
// ==========================================

var incrementAndCheckScript = redis.NewScript(`
local voteKey = KEYS[1]
local roomKey = KEYS[2]
local voteID = ARGV[1]
local ttl = ARGV[2]

-- Increment the vote
local newCount = redis.call("HINCRBY", voteKey, voteID, 1)
redis.call("EXPIRE", voteKey, ttl)

-- Fetch the live client count directly from the room hash
local numClientsStr = redis.call("HGET", roomKey, "client_count")
local numClients = 0
if numClientsStr then
	numClients = tonumber(numClientsStr)
end

-- Compare (Checking for unanimity/100% acceptance)
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
-- Initialize started flag and client_count in the same hash
redis.call("HSET", KEYS[1], "started", "false", "client_count", 0)
-- Set TTL
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

// ==========================================
// Repository
// ==========================================

type RoomRepository struct {
	rdb *redis.Client
}

func NewRoomRepository(rdb *redis.Client) *RoomRepository {
	return &RoomRepository{
		rdb: rdb,
	}
}

func (rr *RoomRepository) IncrementAcceptVote(ctx context.Context, roomID string, voteID string) (bool, error) {
	result, err := incrementAndCheckScript.Run(
		ctx,
		rr.rdb,
		[]string{voteKey(roomID), roomKey(roomID)}, // KEYS[1], KEYS[2]
		voteID,                 // ARGV[1]
		int(roomTTL.Seconds()), // ARGV[2]
	).Int()

	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (rr *RoomRepository) IncrementRoomClientCount(ctx context.Context, roomCode string) (int64, error) {
	count, err := rr.rdb.HIncrBy(ctx, roomKey(roomCode), fieldClientCount, 1).Result()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (rr *RoomRepository) DecrementRoomClientCount(ctx context.Context, roomCode string) (int64, error) {
	count, err := rr.rdb.HIncrBy(ctx, roomKey(roomCode), fieldClientCount, -1).Result()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (rr *RoomRepository) GetRoomClientCount(ctx context.Context, roomCode string) (int64, error) {
	countStr, err := rr.rdb.HGet(ctx, roomKey(roomCode), fieldClientCount).Result()
	if err == redis.Nil {
		return 0, nil // Handle case where client_count field might be missing
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

func (rr *RoomRepository) PublishMajorityFound(ctx context.Context, roomID string, voteID string) error {
	event := RoomEvent{
		Type:   "majority_found",
		Room:   roomID,
		VoteID: voteID,
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return rr.rdb.Publish(ctx, roomEventsChannel, payload).Err()
}

func (rr *RoomRepository) CreateRoom(ctx context.Context, code string) (bool, error) {
	res, err := createRoomScript.Run(
		ctx,
		rr.rdb,
		[]string{roomKey(code)},
		int(roomTTL.Seconds()), // ARGV[1] (TTL in seconds)
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
		return false, fmt.Errorf("room not found")
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
	event := RoomEvent{
		Type: "room_started",
		Room: code,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	res, err := startRoomScript.Run(
		ctx,
		rr.rdb,
		[]string{roomKey(code)},
		roomEventsChannel,
		payload,
	).Int()

	if err != nil {
		return err
	}
	if res == 0 {
		return fmt.Errorf("room not found")
	}

	return nil
}

func (rr *RoomRepository) SubscribeToRoomEvents(ctx context.Context) (*redis.PubSub, error) {
	pubsub := rr.rdb.Subscribe(ctx, roomEventsChannel)

	// ensure subscription is ready
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return nil, err
	}

	return pubsub, nil
}

func (rr *RoomRepository) SetRoomCards(ctx context.Context, code string, cards []util.Card) error {
	data, err := json.Marshal(cards)
	if err != nil {
		return err
	}

	return rr.rdb.HSet(ctx, roomKey(code), fieldCards, data).Err()
}

func (rr *RoomRepository) GetRoomCards(ctx context.Context, code string) ([]util.Card, error) {
	data, err := rr.rdb.HGet(ctx, roomKey(code), fieldCards).Bytes()
	if err == redis.Nil {
		return nil, nil // No cards stored yet
	}
	if err != nil {
		return nil, err
	}

	var cards []util.Card
	if err := json.Unmarshal(data, &cards); err != nil {
		return nil, err
	}

	return cards, nil
}

func (rr *RoomRepository) SetPageToken(ctx context.Context, code string, token string) error {
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
