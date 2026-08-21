package room

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"sync"
	"time"
)

type RoomManager struct {
	mu             sync.RWMutex
	roomRepository Repository
	placesClient   PlacesAPI
	hub            RoomHub
}

// RoomHub is the interface for the WebSocket connection registry.
// Implemented by the ws.Hub to keep room package transport-agnostic.
type RoomHub interface {
	RegisterConn(code string, clientID string) error
	UnregisterConn(code string, clientID string)
	BroadcastRoomStarted(code string)
	BroadcastMajorityFound(code string, voteID string)
	HasRoom(code string) bool
}

func NewRoomManager(repo Repository, placesClient PlacesAPI, hub RoomHub) *RoomManager {
	return &RoomManager{
		roomRepository: repo,
		placesClient:   placesClient,
		hub:            hub,
	}
}

func (rm *RoomManager) HandleVote(ctx context.Context, vote Vote) error {
	if vote.Room == "" {
		return ErrVoteRoomRequired
	}
	if vote.Id == "" {
		return ErrVoteIDRequired
	}
	if vote.ClientID == "" {
		return fmt.Errorf("client id is required")
	}

	// only count ACCEPT votes
	if strings.ToUpper(vote.Result) != "ACCEPT" {
		return nil
	}

	majorityFound, err := rm.roomRepository.IncrementAcceptVote(ctx, vote.Room, vote.Id, vote.ClientID)
	if err != nil {
		return err
	}

	if majorityFound {
		slog.Info("majority found", "room", vote.Room, "voteID", vote.Id)
		rm.roomRepository.PublishMajorityFound(ctx, vote.Room, vote.Id)
	}

	return nil
}

func (rm *RoomManager) StartEventListener(ctx context.Context) {
	for {
		sub, err := rm.roomRepository.SubscribeToRoomEvents(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Error("failed to subscribe to room events", "error", err)
			time.Sleep(2 * time.Second)
			continue
		}

		slog.Info("successfully subscribed to room events")
		eventCh := sub.Events()

		for event := range eventCh {
			rm.handleEvent(event)
		}

		_ = sub.Close()

		if ctx.Err() != nil {
			return
		}
		slog.Warn("room event subscription channel closed unexpectedly, reconnecting in 2 seconds")
		time.Sleep(2 * time.Second)
	}
}

func (rm *RoomManager) handleEvent(event RoomEvent) {
	switch event.Type {
	case "room_started":
		slog.Info("received room_started event", "room", event.Room)
		rm.hub.BroadcastRoomStarted(event.Room)
	case "majority_found":
		slog.Info("received majority_found event", "room", event.Room, "voteID", event.VoteID)
		rm.hub.BroadcastMajorityFound(event.Room, event.VoteID)
	}
}

func (rm *RoomManager) GenerateRoomCode() string {
	// 8 alphanumeric chars from crypto/rand — ~62^8 = ~218 trillion keyspace
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 8)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// fallback to hex if crypto/rand fails (extremely unlikely)
			fallback := make([]byte, 4)
			_, _ = rand.Read(fallback)
			return hex.EncodeToString(fallback)
		}
		result[i] = charset[n.Int64()]
	}
	return string(result)
}

func (rm *RoomManager) AddRoom(ctx context.Context) (string, error) {
	for range 100 { // limit attempts to prevent infinite loop
		code := rm.GenerateRoomCode()
		created, err := rm.roomRepository.CreateRoom(ctx, code)
		if err != nil {
			return "", err
		}
		if created {
			return code, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique room code after 100 attempts")
}

func (rm *RoomManager) StartRoom(ctx context.Context, code string, filters Filters) (bool, error) {
	if err := filters.Validate(); err != nil {
		return false, fmt.Errorf("invalid filters: %w", err)
	}

	cards, nextPageToken, err := rm.placesClient.FetchCards(filters, "")
	if err != nil {
		return false, fmt.Errorf("failed to fetch places: %w", err)
	}

	if len(cards) == 0 {
		return false, ErrNoPlacesFound
	}

	if err := rm.roomRepository.SetRoomCards(ctx, code, cards); err != nil {
		return false, fmt.Errorf("failed to save room cards: %w", err)
	}

	if err := rm.roomRepository.SetPageToken(ctx, code, nextPageToken); err != nil {
		return false, fmt.Errorf("failed to save page token: %w", err)
	}

	err = rm.roomRepository.StartRoom(ctx, code)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (rm *RoomManager) ValidateRoomJoin(ctx context.Context, code string) (bool, error) {
	started, err := rm.roomRepository.CheckIfRoomStarted(ctx, code)
	if err != nil {
		return false, err
	}
	if started {
		return false, nil
	}
	return true, nil
}

func (rm *RoomManager) RegisterConn(ctx context.Context, code string, clientID string) error {
	exists, err := rm.roomRepository.CheckIfRoomExists(ctx, code)
	if err != nil {
		return err
	}
	if !exists {
		return ErrRoomNotFound
	}

	if err := rm.hub.RegisterConn(code, clientID); err != nil {
		return err
	}

	_, err = rm.roomRepository.IncrementRoomClientCount(ctx, code)
	return err
}

func (rm *RoomManager) UnregisterConn(ctx context.Context, code string, clientID string) {
	rm.hub.UnregisterConn(code, clientID)
	_, _ = rm.roomRepository.DecrementRoomClientCount(ctx, code)
}

func (rm *RoomManager) GetRoomCards(ctx context.Context, code string) ([]Card, error) {
	return rm.roomRepository.GetRoomCards(ctx, code)
}

func (rm *RoomManager) GetPhotoURL(photoName string) (string, error) {
	return rm.placesClient.GetPhotoURL(photoName)
}