package room

import (
	"backend/util"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const roomEventsChannel = "room_events"

type RoomRepository struct {
	rdb *redis.Client
}

func NewRoomRepository(rdb *redis.Client) *RoomRepository {
	return &RoomRepository{
		rdb: rdb,
	}
}

var incrementAndCheckScript = redis.NewScript(`
local newCount = redis.call("HINCRBY", KEYS[1], ARGV[1], 1)
local numClients = tonumber(ARGV[2])
if newCount == numClients then
	return 1
else
	return 0
end
`)

func (rr *RoomRepository) IncrementAcceptVote(
	ctx context.Context,
	roomID string,
	voteID string,
	numClients int,
) (bool, error) {

	key := fmt.Sprintf("votes:%s", roomID)

	result, err := incrementAndCheckScript.Run(
		ctx,
		rr.rdb,
		[]string{key}, // KEYS[1]
		voteID,        // ARGV[1]
		numClients,    // ARGV[2]
	).Int()

	if err != nil {
		return false, err
	}

	// result == 1 means threshold reached
	return result == 1, nil
}

func (rr *RoomRepository) IncrementRoomClientCount(
	ctx context.Context,
	roomCode string,
) (int64, error) {
	key := fmt.Sprintf("room:%s:client_count", roomCode)
	count, err := rr.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (rr *RoomRepository) DecrementRoomClientCount(
	ctx context.Context,
	roomCode string,
) (int64, error) {
	key := fmt.Sprintf("room:%s:client_count", roomCode)
	count, err := rr.rdb.Decr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (rr *RoomRepository) GetRoomClientCount(
	ctx context.Context,
	roomCode string,
) (int64, error) {
	key := fmt.Sprintf("room:%s:client_count", roomCode)
	count, err := rr.rdb.Get(ctx, key).Int64()
	if err != nil && err != redis.Nil {
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

func (rr *RoomRepository) CreateRoom(code string) (bool, error) {
	script := `
		if redis.call("EXISTS", KEYS[1]) == 1 then
			return 0
		end
		redis.call("HSET", KEYS[1], "started", "false")
		return 1
	`

	res, err := rr.rdb.Eval(context.Background(), script, []string{"room:" + code}).Int()
	if err != nil {
		return false, err
	}

	return res == 1, nil
}

func (rr *RoomRepository) CheckIfRoomExists(code string) (bool, error) {
	res, err := rr.rdb.Exists(context.Background(), "room:"+code).Result()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

func (rr *RoomRepository) CheckIfRoomStarted(code string) (bool, error) {
	startedStr, err := rr.rdb.HGet(context.Background(), "room:"+code, "started").Result()
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

func (rr *RoomRepository) StartRoom(code string) error {
	event := RoomEvent{
		Type: "room_started",
		Room: code,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	script := `
		if redis.call("EXISTS", KEYS[1]) == 0 then
			return 0
		end
		redis.call("HSET", KEYS[1], "started", "true")
		redis.call("PUBLISH", ARGV[1], ARGV[2])
		return 1
	`

	res, err := rr.rdb.Eval(
		context.Background(),
		script,
		[]string{"room:" + code},
		"room_events",
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

func (rr *RoomRepository) SubscribeToRoomEvents() <-chan *redis.Message {
	pubsub := rr.rdb.Subscribe(context.Background(), "room_events")

	// ensure subscription is ready
	_, err := pubsub.Receive(context.Background())
	if err != nil {
		panic(err)
	}

	return pubsub.Channel()
}

func (rr *RoomRepository) SetRoomCards(ctx context.Context, code string, cards []util.Card) error {
	data, err := json.Marshal(cards)
	if err != nil {
		return err
	}

	key := "room:" + code
	return rr.rdb.HSet(ctx, key, "cards", data).Err()
}

func (rr *RoomRepository) GetRoomCards(ctx context.Context, code string) ([]util.Card, error) {
	key := "room:" + code
	data, err := rr.rdb.HGet(ctx, key, "cards").Bytes()
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
	key := "room:" + code
	return rr.rdb.HSet(ctx, key, "pageToken", token).Err()
}

func (rr *RoomRepository) GetPageToken(ctx context.Context, code string) (string, error) {
	key := "room:" + code
	token, err := rr.rdb.HGet(ctx, key, "pageToken").Result()

	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return token, nil
}
