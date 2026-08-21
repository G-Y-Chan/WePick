package room

type RoomEvent struct {
	Type   string `json:"type"`
	Room   string `json:"room"`
	VoteID string `json:"voteId,omitempty"`
}