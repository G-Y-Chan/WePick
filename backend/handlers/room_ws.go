package handlers

import (
	"net/http"
	"github.com/gorilla/websocket"
	"fmt"
	"backend/util"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// unsafe, but allows all connections. In production, you should check the origin here.
		return true
	},
}

func (s *Server) HandleRoomWS(w http.ResponseWriter, r *http.Request) {
	roomCode := r.URL.Query().Get("roomCode")
	if roomCode == "" {
		http.Error(w, "missing roomCode", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	err = s.RoomManager.RegisterConn(roomCode, conn)
	if err != nil {
		_ = conn.WriteJSON(map[string]any{"type":"ERROR","message": err.Error()})
		_ = conn.Close()
		return
	}

	// Keep alive; when client disconnects, unregister
	for {
		var msg util.Message
		if err := conn.ReadJSON(&msg); err != nil {
			s.RoomManager.UnregisterConn(roomCode, conn)
			_ = conn.Close()
			return
		}

		switch msg.Header {
		case "VOTE_EVENT":
			if msg.VoteObj != nil {
				fmt.Printf("Vote received:\n")
				fmt.Printf("  ID: %s\n", msg.VoteObj.Id)
				fmt.Printf("  Result: %s\n", msg.VoteObj.Result)
			} else {
				fmt.Printf("Empty vote object received.\n")
			}
		default:
			fmt.Println("Wrong header")
		}
	}
}
