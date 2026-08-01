package util

type Vote struct {
	Id     string
	Result string
	Room   string
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
}

type Message struct {
	Header  string  `json:"Header"`
	Body    *string `json:"Body,omitempty"`
	VoteObj *Vote   `json:"VoteObj,omitempty"`
	Cards   []Card  `json:"Cards,omitempty"`
}
