package util

type Vote struct {
	Id     string
	Result string
}

type Message struct {
	Header  string  `json:"Header"`
	Body    *string `json:"Body,omitempty"`
	VoteObj *Vote   `json:"VoteObj,omitempty"`
}
