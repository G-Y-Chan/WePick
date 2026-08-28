package ws

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"backend/internal/adapters/wshub"
	"backend/internal/domain/room"
)

// Handler is the WebSocket transport layer. It depends only on room.Service
// and the wshub connection adapter; it has no knowledge of Redis or Google.
type Handler struct {
	svc      room.Service
	upgrader websocket.Upgrader
	logger   *slog.Logger
}

// NewHandler constructs the WebSocket transport Handler. allowedOrigins uses
// the same allow-list policy as the HTTP CORS middleware for WebSocket upgrade
// requests.
func NewHandler(svc room.Service, allowedOrigins []string, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return &Handler{
		svc:    svc,
		logger: logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return originAllowed(r.Header.Get("Origin"), allowedOrigins)
			},
		},
	}
}

// ServeWS handles GET /ws?roomCode=... .
//
// It upgrades the connection, wraps it in a wshub.Conn whose WritePump is the
// only goroutine allowed to write to the underlying websocket.Conn, calls
// svc.Connect, then hands inbound frames to wshub.Conn.ReadLoop with a callback
// that delegates votes to svc.SubmitVote. On read-loop termination the client
// is disconnected using r.Context() (fixing the legacy dropped-context bug) and
// the connection is closed.
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	rawCode := r.URL.Query().Get("roomCode")
	if rawCode == "" {
		http.Error(w, "missing roomCode", http.StatusBadRequest)
		return
	}

	code, err := room.ParseCode(rawCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wsConn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade already wrote the HTTP error response when appropriate.
		return
	}

	conn := wshub.NewConn(wsConn, newConnectionID(), 32)
	go conn.WritePump()

	if err := h.svc.Connect(r.Context(), code, conn); err != nil {
		h.logger.Warn("websocket connect failed", "room_code", code.String(), "err", err)
		_ = conn.Send(room.OutboundEvent{
			Type:    room.OutboundError,
			Message: err.Error(),
		})
		// Give WritePump a brief window to flush the error frame before the
		// connection is closed. The close path remains safe under the Conn
		// single-writer concurrency contract.
		time.Sleep(5 * time.Millisecond)
		_ = conn.Close()
		return
	}

	h.logger.Info("websocket client connected", "room_code", code.String(), "conn_id", conn.ID())

	readErr := conn.ReadLoop(func(vote room.Vote) error {
		return h.svc.SubmitVote(r.Context(), vote)
	})

	h.svc.Disconnect(r.Context(), code, conn)
	_ = conn.Close()

	if readErr != nil {
		h.logger.Info("websocket client disconnected",
			"room_code", code.String(),
			"conn_id", conn.ID(),
			"err", readErr,
		)
	}
}

// originAllowed reports whether the WebSocket Origin header matches one of the
// configured allow-list entries. Exact entries compare verbatim; "*." entries
// match any subdomain suffix.
func originAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		if strings.HasPrefix(allowed, "*.") {
			suffix := strings.TrimPrefix(allowed, "*")
			if strings.HasSuffix(origin, suffix) {
				return true
			}
			continue
		}

		if origin == allowed {
			return true
		}
	}
	return false
}

func newConnectionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err == nil {
		return hex.EncodeToString(b)
	}
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
