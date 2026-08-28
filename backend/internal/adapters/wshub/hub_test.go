package wshub

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"testing"

	"backend/internal/domain/room"
)

type fakeClient struct {
	mu     sync.Mutex
	id     string
	events []room.OutboundEvent
}

func newFakeClient(id string) *fakeClient {
	return &fakeClient{id: id}
}

func (f *fakeClient) ID() string {
	return f.id
}

func (f *fakeClient) Send(evt room.OutboundEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, evt)
	return nil
}

func (f *fakeClient) eventCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.events)
}

type failingClient struct {
	*fakeClient
}

func (f *failingClient) Send(room.OutboundEvent) error {
	return errors.New("send buffer full")
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestHubBroadcastConcurrentWithDirectSendIsRaceFree(t *testing.T) {
	h := NewHub(testLogger())

	const numClients = 16
	const iterations = 1000

	clients := make([]*fakeClient, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = newFakeClient(fmt.Sprintf("conn-%d", i))
		if err := h.Register("123456", clients[i]); err != nil {
			t.Fatalf("Register(%d): %v", i, err)
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			h.BroadcastRoomStarted("123456")
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = clients[0].Send(room.OutboundEvent{
				Type:   room.OutboundMajorityFound,
				CardID: "card-1",
			})
		}
	}()

	wg.Wait()

	if got := clients[0].eventCount(); got != iterations*2 {
		t.Fatalf("client 0 received %d events, want %d", got, iterations*2)
	}
}

func TestHubUnregisterRemovesEmptyRoom(t *testing.T) {
	h := NewHub(testLogger())

	first := newFakeClient("conn-1")
	second := newFakeClient("conn-2")

	if err := h.Register("123456", first); err != nil {
		t.Fatalf("Register(first): %v", err)
	}
	if err := h.Register("123456", second); err != nil {
		t.Fatalf("Register(second): %v", err)
	}

	h.Unregister("123456", first)

	h.mu.RLock()
	_, exists := h.rooms["123456"]
	h.mu.RUnlock()
	if !exists {
		t.Fatal("room should still exist while one client remains connected")
	}

	h.Unregister("123456", second)

	h.mu.RLock()
	_, exists = h.rooms["123456"]
	h.mu.RUnlock()
	if exists {
		t.Fatal("room entry should be removed once its last client disconnects")
	}
}

func TestHubBroadcastEvictsSlowClient(t *testing.T) {
	h := NewHub(testLogger())

	healthy := newFakeClient("conn-healthy")
	slow := &failingClient{fakeClient: newFakeClient("conn-slow")}

	if err := h.Register("123456", healthy); err != nil {
		t.Fatalf("Register(healthy): %v", err)
	}
	if err := h.Register("123456", slow); err != nil {
		t.Fatalf("Register(slow): %v", err)
	}

	h.BroadcastMajorityFound("123456", "card-1")

	h.mu.RLock()
	_, slowStillRegistered := h.rooms["123456"][slow]
	_, healthyStillRegistered := h.rooms["123456"][healthy]
	h.mu.RUnlock()

	if slowStillRegistered {
		t.Fatal("slow client should have been evicted after a failed Send")
	}
	if !healthyStillRegistered {
		t.Fatal("healthy client should remain registered")
	}
	if got := healthy.eventCount(); got != 1 {
		t.Fatalf("healthy client received %d events, want 1", got)
	}
}

func TestHubRegisterValidation(t *testing.T) {
	h := NewHub(testLogger())

	if err := h.Register("", newFakeClient("conn-1")); err == nil {
		t.Fatal("expected error for empty room code")
	}
	if err := h.Register("123456", nil); err == nil {
		t.Fatal("expected error for nil client")
	}
}
