package redisrepo

import "github.com/redis/go-redis/v9"

// Lua scripts carried over verbatim from room/room_repository.go.
// The script bodies, Redis key layout, hash field names, pub/sub channel name,
// and the 24h TTL are preserved unchanged — this phase relocates and re-interfaces
// existing logic, it does not redesign the Redis schema.

const (
	// roomEventsChannel is the Redis Pub/Sub channel name for room-level events.
	// Must match the channel name used in the legacy code.
	roomEventsChannel = "room_events"

	// RoomTTL is the default time-to-live for room keys in Redis.
	RoomTTL = 24 * 3600 // 24 hours in seconds, matching the legacy roomTTL constant.
)

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