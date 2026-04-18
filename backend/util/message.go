package util

type Vote struct {
	Id     string
	Result string
	Room   string
}

type Card struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Message struct {
	Header  string  `json:"Header"`
	Body    *string `json:"Body,omitempty"`
	VoteObj *Vote   `json:"VoteObj,omitempty"`
	Cards   []Card  `json:"Cards,omitempty"`
}
