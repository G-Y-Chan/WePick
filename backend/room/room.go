package room

type Room struct {
    Code      string
    Started   bool
}

func NewRoom(code string) *Room {
	return &Room{
		Code: code,
		Started: false,
	}
}
