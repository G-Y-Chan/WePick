package room

// Event is the internal pub/sub bus contract (Redis Pub/Sub payload today,
// but the domain layer only knows about "Event", not "Redis").
type EventType string

const (
	EventRoomStarted   EventType = "room_started"
	EventMajorityFound EventType = "majority_found"
)

// Event represents a cross-process bus message.
type Event struct {
	Type   EventType
	Room   Code
	CardID string
}
