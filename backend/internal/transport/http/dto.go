package api

// These DTOs are the wire format for HTTP JSON responses and WebSocket messages.
// They are intentionally separate from the domain types in internal/room to
// enforce a clean translation boundary.

type ErrorResponse struct {
	Header  string `json:"Header"`
	Body    string `json:"Body"`
	Message string `json:"Message,omitempty"`
}

type Message struct {
	Header  string      `json:"Header"`
	Body    *string     `json:"Body,omitempty"`
	VoteObj *Vote       `json:"VoteObj,omitempty"`
	Cards   []Card      `json:"Cards,omitempty"`
}

type Vote struct {
	Id     string `json:"Id"`
	Result string `json:"Result"`
	Room   string `json:"Room"`
}

type Card struct {
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

type StartRoomPayload struct {
	Filters struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Radius    int     `json:"radius"`
		MaxPrice  int     `json:"maxPrice"`
		Category  string  `json:"category"`
		OpenNow   bool    `json:"openNow"`
	} `json:"filters"`
}