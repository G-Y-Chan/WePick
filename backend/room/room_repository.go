package room

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"context"
	"strconv"
	"encoding/json"
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
if newCount == tonumber(ARGV[2]) then
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
		[]string{key},   // KEYS[1]
		voteID,          // ARGV[1]
		numClients,      // ARGV[2]
	).Int()

	if err != nil {
		return false, err
	}

	// result == 1 means threshold reached
	return result == 1, nil
}

func (rr *RoomRepository) PublishMajorityFound(ctx context.Context, roomID string, voteID string) error {
	event := RoomEvent{
		Type: "majority_found",
		Room: roomID,
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

	res, err := rr.rdb.Eval(context.Background(), script, []string{"room:"+code}).Int()
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
		[]string{"room:"+code},
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
