package ws

import (
	"backend/internal/room"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Actual origin checking is done via middleware; WS upgrade
		// inherits the same security boundary.
		return true
	},
}

// Handler holds dependencies for the WebSocket handler.
type Handler struct {
	RoomManager *room.RoomManager
	Hub         *Hub
}

func NewHandler(rm *room.RoomManager, hub *Hub) *Handler {
	return &Handler{
		RoomManager: rm,
		Hub:         hub,
	}
}

func generateClientID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (h *Handler) HandleRoomWS(w http.ResponseWriter, r *http.Request) {
	roomCode := r.URL.Query().Get("roomCode")
	if roomCode == "" {
		http.Error(w, "missing roomCode", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("websocket upgrade failed", "error", err)
		return
	}

	clientID := generateClientID()
	slog.Info("new websocket connection", "clientID", clientID, "room", roomCode)

	// Create the client
	client := &Client{
		ID:       clientID,
		conn:     conn,
		send:     make(chan []byte, 256),
		hub:      h.Hub,
		roomCode: roomCode,
	}

	// Register with the room manager (validates room exists, increments count)
	if err := h.RoomManager.RegisterConn(r.Context(), roomCode, clientID); err != nil {
		slog.Warn("room registration failed", "clientID", clientID, "room", roomCode, "error", err)
		_ = conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, err.Error()),
			time.Now().Add(time.Second))
		_ = conn.Close()
		return
	}

	// Add to hub
	h.Hub.AddClient(roomCode, client)

	// Start write pump
	go client.writePump()

	// Read loop with pong handling
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg struct {
			Header  string     `json:"Header"`
			VoteObj *room.Vote `json:"VoteObj,omitempty"`
		}

		if err := conn.ReadJSON(&msg); err != nil {
			slog.Info("client disconnected", "clientID", clientID, "room", roomCode, "error", err)
			break
		}

		switch msg.Header {
		case "VOTE_EVENT":
			if msg.VoteObj != nil {
				msg.VoteObj.ClientID = clientID
				msg.VoteObj.Room = roomCode

				if strings.ToUpper(msg.VoteObj.Result) == "ACCEPT" {
					slog.Info("vote received",
						"clientID", clientID,
						"voteID", msg.VoteObj.Id,
						"result", msg.VoteObj.Result,
						"room", roomCode,
					)

					if err := h.RoomManager.HandleVote(context.Background(), *msg.VoteObj); err != nil {
						slog.Warn("vote handling failed", "error", err)
						errMsg, _ := json.Marshal(map[string]string{"error": err.Error()})
						client.send <- errMsg
					}
				}
			} else {
				slog.Warn("empty vote object received", "clientID", clientID)
			}
		default:
			slog.Warn("unknown message header", "header", msg.Header, "clientID", clientID)
		}
	}

	// Cleanup
	h.Hub.RemoveClient(roomCode, clientID)
	h.RoomManager.UnregisterConn(context.Background(), roomCode, clientID)
	_ = conn.Close()
}