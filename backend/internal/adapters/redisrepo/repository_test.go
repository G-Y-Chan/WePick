package redisrepo

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"backend/internal/domain/room"
)

// newTestRepo returns a Repository backed by a miniredis instance and the
// miniredis handle so callers can manipulate time or inspect internals.
func newTestRepo(t *testing.T) (*Repository, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return New(rdb), mr
}

// ---------------------------------------------------------------------------
// CreateRoom
// ---------------------------------------------------------------------------

func TestCreateRoom_Success(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	code := room.Code("123456")
	created, err := repo.CreateRoom(ctx, code, 24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Fatal("expected room to be created")
	}

	// Verify the key actually exists.
	exists, err := repo.Exists(ctx, code)
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !exists {
		t.Fatal("room should exist after CreateRoom")
	}
}

func TestCreateRoom_Collision(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	code := room.Code("111111")
	created, err := repo.CreateRoom(ctx, code, 24*time.Hour)
	if err != nil {
		t.Fatalf("first create: %v", err)
	}
	if !created {
		t.Fatal("first create should succeed")
	}

	// Second create with same code must report collision.
	created, err = repo.CreateRoom(ctx, code, 24*time.Hour)
	if err != nil {
		t.Fatalf("second create: %v", err)
	}
	if created {
		t.Fatal("second create should report collision")
	}
}

func TestCreateRoom_SetsStartedFalseAndClientCountZero(t *testing.T) {
	repo, mr := newTestRepo(t)
	ctx := context.Background()

	code := room.Code("222222")
	_, err := repo.CreateRoom(ctx, code, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}

	started, err := repo.IsStarted(ctx, code)
	if err != nil {
		t.Fatalf("IsStarted: %v", err)
	}
	if started {
		t.Fatal("new room must not be started")
	}

	count, err := repo.ClientCount(ctx, code)
	if err != nil {
		t.Fatalf("ClientCount: %v", err)
	}
	if count != 0 {
		t.Fatalf("new room client_count should be 0, got %d", count)
	}

	// Also check TTL was set.
	key := roomKey(code)
	ttl := mr.TTL(key)
	if ttl <= 0 {
		t.Fatalf("TTL on room key should be positive, got %v", ttl)
	}
}

// ---------------------------------------------------------------------------
// IsStarted / MarkStarted idempotency
// ---------------------------------------------------------------------------

func TestIsStarted_NotFound(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	_, err := repo.IsStarted(ctx, room.Code("nope__"))
	if err == nil {
		t.Fatal("expected error for non-existent room")
	}
	if err != room.ErrRoomNotFound {
		t.Fatalf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestMarkStarted_Success(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	code := room.Code("333333")
	_, err := repo.CreateRoom(ctx, code, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}

	err = repo.MarkStarted(ctx, code)
	if err != nil {
		t.Fatalf("MarkStarted: %v", err)
	}

	started, err := repo.IsStarted(ctx, code)
	if err != nil {
		t.Fatalf("IsStarted: %v", err)
	}
	if !started {
		t.Fatal("room should be started after MarkStarted")
	}
}

func TestMarkStarted_NotFound(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	err := repo.MarkStarted(ctx, room.Code("999999"))
	if err == nil {
		t.Fatal("expected error for non-existent room")
	}
	if err != room.ErrRoomNotFound {
		t.Fatalf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestMarkStarted_Idempotent(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	code := room.Code("444444")
	_, err := repo.CreateRoom(ctx, code, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}

	// First call.
	if err := repo.MarkStarted(ctx, code); err != nil {
		t.Fatalf("first MarkStarted: %v", err)
	}
	// Second call should succeed (it just sets started=true again).
	if err := repo.MarkStarted(ctx, code); err != nil {
		t.Fatalf("second MarkStarted should not fail: %v", err)
	}

	started, err := repo.IsStarted(ctx, code)
	if err != nil {
		t.Fatalf("IsStarted: %v", err)
	}
	if !started {
		t.Fatal("room should still be started")
	}
}

// ---------------------------------------------------------------------------
// Cards round-tripping (SaveCards → LoadCards)
// ---------------------------------------------------------------------------

func TestSaveLoadCards_RoundTrip(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	code := room.Code("555555")
	_, err := repo.CreateRoom(ctx, code, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}

	original := []room.Card{
		{
			ID:          "place_1",
			Title:       "Test Place",
			Category:    "restaurant",
			PriceLevel:  "PRICE_LEVEL_MODERATE",
			Rating:      4.3,
			ReviewCount: 120,
			OpenNow:     true,
			Summary:     "A nice place",
			Address:     "123 Test St",
			PhotoRef:    "photo_ref_1",
		},
		{
			ID:          "place_2",
			Title:       "Another Place",
			Category:    "cafe",
			PriceLevel:  "PRICE_LEVEL_INEXPENSIVE",
			Rating:      3.9,
			ReviewCount: 45,
			OpenNow:     false,
			Summary:     "",
			Address:     "456 Another Ave",
			PhotoRef:    "photo_ref_2",
		},
	}

	if err := repo.SaveCards(ctx, code, original); err != nil {
		t.Fatalf("SaveCards: %v", err)
	}

	loaded, err := repo.LoadCards(ctx, code)
	if err != nil {
		t.Fatalf("LoadCards: %v", err)
	}

	if len(loaded) != len(original) {
		t.Fatalf("expected %d cards, got %d", len(original), len(loaded))
	}

	for i := range original {
		if original[i].ID != loaded[i].ID {
			t.Errorf("card[%d].ID: want %q, got %q", i, original[i].ID, loaded[i].ID)
		}
		if original[i].PhotoRef != loaded[i].PhotoRef {
			t.Errorf("card[%d].PhotoRef: want %q, got %q", i, original[i].PhotoRef, loaded[i].PhotoRef)
		}
	}
}

func TestLoadCards_Empty(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	code := room.Code("666666")
	_, err := repo.CreateRoom(ctx, code, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}

	cards, err := repo.LoadCards(ctx, code)
	if err != nil {
		t.Fatalf("LoadCards: %v", err)
	}
	if cards != nil {
		t.Fatalf("expected nil for unset cards, got %v", cards)
	}
}

func TestSaveCards_PreservesLegacyJSONFormat(t *testing.T) {
	// Verify the stored JSON uses "photoName" (legacy key) not "PhotoRef".
	repo, mr := newTestRepo(t)
	ctx := context.Background()

	code := room.Code("777777")
	_, err := repo.CreateRoom(ctx, code, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}

	cards := []room.Card{
		{ID: "p1", PhotoRef: "ref123"},
	}
	if err := repo.SaveCards(ctx, code, cards); err != nil {
		t.Fatalf("SaveCards: %v", err)
	}

	// Read the raw JSON from Redis to confirm "photoName" key.
	raw := mr.HGet(roomKey(code), "cards")

	var parsed []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	if len(parsed) != 1 {
		t.Fatalf("expected 1 card in raw JSON, got %d", len(parsed))
	}
	if _, ok := parsed[0]["photoName"]; !ok {
		t.Fatalf("raw JSON missing 'photoName' key; got keys: %v", parsed[0])
	}
	if _, ok := parsed[0]["PhotoRef"]; ok {
		t.Fatalf("raw JSON should not contain 'PhotoRef' key")
	}
	if parsed[0]["photoName"] != "ref123" {
		t.Fatalf("photoName: want %q, got %v", "ref123", parsed[0]["photoName"])
	}
}

// ---------------------------------------------------------------------------
// Page token round-tripping
// ---------------------------------------------------------------------------

func TestSaveLoadPageToken_RoundTrip(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	code := room.Code("888888")
	_, err := repo.CreateRoom(ctx, code, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}

	token := "nextPageToken_abc123"
	if err := repo.SavePageToken(ctx, code, token); err != nil {
		t.Fatalf("SavePageToken: %v", err)
	}

	loaded, err := repo.LoadPageToken(ctx, code)
	if err != nil {
		t.Fatalf("LoadPageToken: %v", err)
	}
	if loaded != token {
		t.Fatalf("page token: want %q, got %q", token, loaded)
	}
}

func TestLoadPageToken_Empty(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	code := room.Code("135791")
	_, err := repo.CreateRoom(ctx, code, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}

	token, err := repo.LoadPageToken(ctx, code)
	if err != nil {
		t.Fatalf("LoadPageToken: %v", err)
	}
	if token != "" {
		t.Fatalf("expected empty token, got %q", token)
	}
}

// ---------------------------------------------------------------------------
// Client count increment / decrement
// ---------------------------------------------------------------------------

func TestClientCount_IncDec(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	code := room.Code("246800")
	_, err := repo.CreateRoom(ctx, code, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}

	// Start at 0.
	c, err := repo.ClientCount(ctx, code)
	if err != nil {
		t.Fatalf("ClientCount: %v", err)
	}
	if c != 0 {
		t.Fatalf("initial client_count: want 0, got %d", c)
	}

	// Inc a few times.
	for i := int64(1); i <= 3; i++ {
		got, err := repo.IncrClientCount(ctx, code)
		if err != nil {
			t.Fatalf("IncrClientCount #%d: %v", i, err)
		}
		if got != i {
			t.Fatalf("IncrClientCount #%d: want %d, got %d", i, i, got)
		}
	}

	// Dec back to 0.
	for i := int64(2); i >= 0; i-- {
		got, err := repo.DecrClientCount(ctx, code)
		if err != nil {
			t.Fatalf("DecrClientCount: %v", err)
		}
		if got != i {
			t.Fatalf("DecrClientCount: want %d, got %d", i, got)
		}
	}
}

// ---------------------------------------------------------------------------
// RegisterAcceptVote — majority Lua script
// ---------------------------------------------------------------------------

func TestRegisterAcceptVote_MajorityReached(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	code := room.Code("112233")
	_, err := repo.CreateRoom(ctx, code, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}

	// Simulate 2 connected clients.
	if _, err := repo.IncrClientCount(ctx, code); err != nil {
		t.Fatalf("IncrClientCount: %v", err)
	}
	if _, err := repo.IncrClientCount(ctx, code); err != nil {
		t.Fatalf("IncrClientCount: %v", err)
	}

	cardID := "card_abc"

	// First vote: 1/2, no majority.
	maj, err := repo.RegisterAcceptVote(ctx, code, cardID)
	if err != nil {
		t.Fatalf("RegisterAcceptVote #1: %v", err)
	}
	if maj {
		t.Fatal("should not be majority after 1/2 votes")
	}

	// Second vote: 2/2, majority reached.
	maj, err = repo.RegisterAcceptVote(ctx, code, cardID)
	if err != nil {
		t.Fatalf("RegisterAcceptVote #2: %v", err)
	}
	if !maj {
		t.Fatal("should be majority after 2/2 votes")
	}
}

func TestRegisterAcceptVote_NoMajorityWithSingleClient(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	code := room.Code("445566")
	_, err := repo.CreateRoom(ctx, code, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}

	// 1 client.
	if _, err := repo.IncrClientCount(ctx, code); err != nil {
		t.Fatalf("IncrClientCount: %v", err)
	}

	cardID := "card_solo"
	maj, err := repo.RegisterAcceptVote(ctx, code, cardID)
	if err != nil {
		t.Fatalf("RegisterAcceptVote: %v", err)
	}
	if !maj {
		t.Fatal("single client should reach majority immediately")
	}
}

func TestRegisterAcceptVote_ZeroClients(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	code := room.Code("778899")
	_, err := repo.CreateRoom(ctx, code, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}
	// client_count = 0 (default).

	cardID := "card_zero"
	maj, err := repo.RegisterAcceptVote(ctx, code, cardID)
	if err != nil {
		t.Fatalf("RegisterAcceptVote: %v", err)
	}
	// With numClients=0, the Lua script returns 0 (majority not reached).
	if maj {
		t.Fatal("should not be majority when client_count is 0")
	}
}

// ---------------------------------------------------------------------------
// PublishEvent / SubscribeEvents
// ---------------------------------------------------------------------------

func TestPublishAndSubscribe(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	code := room.Code("998877")
	_, err := repo.CreateRoom(ctx, code, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}

	sub, err := repo.SubscribeEvents(ctx)
	if err != nil {
		t.Fatalf("SubscribeEvents: %v", err)
	}
	defer sub.Close()

	// Publish a majority_found event.
	sent := room.Event{
		Type:   room.EventMajorityFound,
		Room:   code,
		CardID: "card_majority",
	}
	if err := repo.PublishEvent(ctx, sent); err != nil {
		t.Fatalf("PublishEvent: %v", err)
	}

	// Wait for the event.
	select {
	case evt := <-sub.Events():
		if evt.Type != room.EventMajorityFound {
			t.Errorf("event type: want %s, got %s", room.EventMajorityFound, evt.Type)
		}
		if evt.Room != code {
			t.Errorf("event room: want %s, got %s", code, evt.Room)
		}
		if evt.CardID != "card_majority" {
			t.Errorf("event CardID: want %q, got %q", "card_majority", evt.CardID)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for event")
	}
}

func TestMarkStarted_PublishesEvent(t *testing.T) {
	// Verify that MarkStarted publishes a room_started event via the Lua script.
	repo, _ := newTestRepo(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	code := room.Code("665544")
	_, err := repo.CreateRoom(ctx, code, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}

	sub, err := repo.SubscribeEvents(ctx)
	if err != nil {
		t.Fatalf("SubscribeEvents: %v", err)
	}
	defer sub.Close()

	if err := repo.MarkStarted(ctx, code); err != nil {
		t.Fatalf("MarkStarted: %v", err)
	}

	select {
	case evt := <-sub.Events():
		if evt.Type != room.EventRoomStarted {
			t.Errorf("event type: want %s, got %s", room.EventRoomStarted, evt.Type)
		}
		if evt.Room != code {
			t.Errorf("event room: want %s, got %s", code, evt.Room)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for room_started event")
	}
}

// ---------------------------------------------------------------------------
// PublishEvent wire format compatibility
// ---------------------------------------------------------------------------

func TestPublishEvent_WireFormatCompatibility(t *testing.T) {
	// Assert that the published JSON matches the legacy RoomEvent shape.
	repo, mr := newTestRepo(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	code := room.Code("112211")
	_, err := repo.CreateRoom(ctx, code, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}

	// Use miniredis's direct Subscriber to capture the raw payload.
	// Start reading from Messages() in a goroutine BEFORE publishing,
	// because miniredis delivers synchronously to an unbuffered channel.
	sub := mr.NewSubscriber()
	sub.Subscribe(roomEventsChannel)

	msgCh := make(chan miniredis.PubsubMessage, 1)
	go func() {
		msg, ok := <-sub.Messages()
		if ok {
			msgCh <- msg
		}
	}()

	evt := room.Event{
		Type:   room.EventMajorityFound,
		Room:   code,
		CardID: "card_legacy",
	}
	if err := repo.PublishEvent(ctx, evt); err != nil {
		t.Fatalf("PublishEvent: %v", err)
	}

	// Wait for the message.
	select {
	case msg := <-msgCh:
		if msg.Channel != roomEventsChannel {
			t.Fatalf("channel: want %q, got %q", roomEventsChannel, msg.Channel)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(msg.Message), &parsed); err != nil {
			t.Fatalf("unmarshal event: %v", err)
		}

		// Legacy format uses "voteId", not "cardID".
		if _, ok := parsed["voteId"]; !ok {
			t.Fatalf("published event missing 'voteId' key; got keys: %v", parsed)
		}
		if parsed["voteId"] != "card_legacy" {
			t.Fatalf("voteId: want %q, got %v", "card_legacy", parsed["voteId"])
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for pubsub message")
	}
}
