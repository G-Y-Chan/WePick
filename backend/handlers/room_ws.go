package handlers

import (
	"net/http"
	"strconv"
	"github.com/gorilla/websocket"
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

	started, err := srv.RoomService.RegisterConn(roomCode, conn)
	if err != nil {
		_ = conn.WriteJSON(map[string]any{"type":"ERROR","message": err.Error()})
		_ = conn.Close()
		return
	}

	_ = conn.WriteJSON(map[string]any{"type":"JOINED"})

	// If room already started, tell them immediately
	if started {
		_ = conn.WriteJSON(map[string]any{"type":"START"})
	}

	// Keep alive; when client disconnects, unregister
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			srv.RoomService.UnregisterConn(roomCode, conn)
			_ = conn.Close()
			return
		}
	}
}
