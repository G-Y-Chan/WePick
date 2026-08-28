package wshub

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"backend/internal/domain/room"
)

const (
	// writeWait is the maximum time allowed for a single WebSocket write.
	writeWait = 10 * time.Second

	// pongWait is how long the connection is allowed to go without receiving
	// a pong before the read loop treats the connection as dead.
	pongWait = 60 * time.Second

	// pingPeriod must be shorter than pongWait so a ping is sent before the
	// client is expected to answer the previous one.
	pingPeriod = (pongWait * 9) / 10

	// maxMessageSize caps inbound WebSocket frames.
	maxMessageSize = 4096
)

var (
	// ErrConnClosed is returned by Send once the connection has been closed.
	ErrConnClosed = errors.New("websocket connection closed")

	// ErrSendBufferFull is returned by Send when the client is a slow consumer.
	ErrSendBufferFull = errors.New("websocket send buffer full")
)

// wireMessage mirrors the legacy util.Message envelope exactly. Clients depend
// on the capitalized JSON keys and on the untagged vote object fields.
type wireMessage struct {
	Header  string    `json:"Header"`
	Body    *string   `json:"Body,omitempty"`
	VoteObj *wireVote `json:"VoteObj,omitempty"`
	Cards   []any     `json:"Cards,omitempty"`
}

// wireVote mirrors util.Vote exactly. It intentionally has no json tags, so
// the wire keys are the bare Go field names "Id"/"Result"/"Room".
type wireVote struct {
	Id     string
	Result string
	Room   string
}

// wireError mirrors the legacy inline map {"type": "ERROR", "message": "..."}
// used when rejecting a WebSocket connection or reporting a vote error.
type wireError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Conn implements room.Client and wraps a single *websocket.Conn.
//
// CONCURRENCY CONTRACT: no goroutine other than Conn's own WritePump goroutine
// may ever call a write method on the underlying *websocket.Conn. All outbound
// traffic funnels through Send's internal channel. This is the direct fix for
// the data race identified in the architectural audit.
type Conn struct {
	ws   *websocket.Conn
	id   string
	send chan []byte
	done chan struct{}

	closeOnce sync.Once
	writeMu   sync.Mutex
}

// NewConn creates a Conn wrapping ws. sendBuffer controls how many outbound
// messages may be queued before Send reports a slow consumer.
func NewConn(ws *websocket.Conn, id string, sendBuffer int) *Conn {
	if sendBuffer <= 0 {
		sendBuffer = 32
	}

	return &Conn{
		ws:   ws,
		id:   id,
		send: make(chan []byte, sendBuffer),
		done: make(chan struct{}),
	}
}

// ID returns the connection identifier.
func (c *Conn) ID() string {
	return c.id
}

// Send queues an outbound event for delivery by WritePump.
// It is non-blocking: a full buffer returns ErrSendBufferFull so the caller can
// evict the slow client instead of stalling every other client's broadcast.
func (c *Conn) Send(evt room.OutboundEvent) error {
	payload, err := marshalOutbound(evt)
	if err != nil {
		return err
	}

	select {
	case <-c.done:
		return ErrConnClosed
	default:
	}

	select {
	case c.send <- payload:
		return nil
	default:
		return ErrSendBufferFull
	}
}

// Close signals the connection to shut down and closes the underlying
// *websocket.Conn under the same write mutex used by WritePump, so a Close call
// can never race with an in-flight WriteMessage.
func (c *Conn) Close() error {
	c.closeOnce.Do(func() {
		close(c.done)
	})

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if c.ws == nil {
		return ErrConnClosed
	}

	_ = c.ws.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	return c.ws.Close()
}

// WritePump is the ONLY goroutine allowed to write to the underlying
// *websocket.Conn. It drains c.send and also sends periodic pings so dead
// peers are detected by ReadLoop.
func (c *Conn) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	defer c.closeUnderlying()

	for {
		select {
		case payload := <-c.send:
			if err := c.write(websocket.TextMessage, payload); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.done:
			_ = c.write(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			)
			return
		}
	}
}

// ReadLoop reads inbound frames from the WebSocket. It never writes to the
// underlying connection directly. Inbound vote frames are decoded into
// room.Vote and dispatched to onVote. When onVote reports an error, the error
// is sent back to the client via the safe Send path, matching legacy behavior.
func (c *Conn) ReadLoop(onVote func(room.Vote) error) error {
	if c.ws == nil {
		return ErrConnClosed
	}

	c.ws.SetReadLimit(maxMessageSize)
	_ = c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		return c.ws.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		var msg wireMessage
		if err := c.ws.ReadJSON(&msg); err != nil {
			return err
		}

		if msg.Header != "VOTE_EVENT" || msg.VoteObj == nil {
			continue
		}

		vote := room.Vote{
			Room:   room.Code(msg.VoteObj.Room),
			CardID: msg.VoteObj.Id,
			Result: normalizeVoteResult(msg.VoteObj.Result),
		}

		if err := onVote(vote); err != nil {
			_ = c.Send(room.OutboundEvent{
				Type:    room.OutboundError,
				Message: err.Error(),
			})
		}
	}
}

// normalizeVoteResult preserves the legacy behavior where any value other than
// ACCEPT is silently treated as a non-accept vote rather than rejected.
func normalizeVoteResult(raw string) room.VoteResult {
	if parsed, err := room.ParseVoteResult(raw); err == nil {
		return parsed
	}
	return room.VoteReject
}

// write serializes all calls to websocket.Conn.WriteMessage. WritePump is the
// only caller in normal operation, but Close also takes writeMu so Close never
// overlaps a write.
func (c *Conn) write(messageType int, payload []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if c.ws == nil {
		return ErrConnClosed
	}

	return c.ws.WriteMessage(messageType, payload)
}

// closeUnderlying closes the connection under the write mutex. It is safe to
// call even if Close already closed the connection.
func (c *Conn) closeUnderlying() {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if c.ws != nil {
		_ = c.ws.Close()
	}
}

// marshalOutbound converts the transport-agnostic room.OutboundEvent into the
// legacy wire envelope expected by existing clients.
func marshalOutbound(evt room.OutboundEvent) ([]byte, error) {
	switch evt.Type {
	case room.OutboundStart:
		return json.Marshal(wireMessage{Header: "START"})
	case room.OutboundMajorityFound:
		return json.Marshal(wireMessage{
			Header:  "MAJORITY_FOUND",
			VoteObj: &wireVote{Id: evt.CardID},
		})
	case room.OutboundError:
		return json.Marshal(wireError{
			Type:    "ERROR",
			Message: evt.Message,
		})
	default:
		return nil, fmt.Errorf("unknown outbound event type %q", evt.Type)
	}
}

// Compile-time contract check.
var _ room.Client = (*Conn)(nil)
