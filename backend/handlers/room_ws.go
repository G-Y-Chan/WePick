package handlers

import (
	"net/http"
	"strconv"
	"github.com/gorilla/websocket"
	"fmt"
	"backend/util"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (srv *Server) HandleRoomWS(w http.ResponseWriter, r *http.Request) {
	roomCodeStr := r.URL.Query().Get("roomCode")
	if roomCodeStr == "" {
		http.Error(w, "missing roomCode", http.StatusBadRequest)
		return
	}
	roomCode, err := strconv.Atoi(roomCodeStr)
	if err != nil {
		http.Error(w, "invalid roomCode", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	_, err = srv.RoomService.RegisterConn(roomCode, conn)
	if err != nil {
		_ = conn.WriteJSON(map[string]any{"type":"ERROR","message": err.Error()})
		_ = conn.Close()
		return
	}

	// Keep alive; when client disconnects, unregister
	for {
		var msg util.Message
		if err := conn.ReadJSON(&msg); err != nil {
			srv.RoomService.UnregisterConn(roomCode, conn)
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
		}
	}
}
