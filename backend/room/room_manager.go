package room

import (
	"backend/services"
	"backend/util"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type RoomManager struct {
	mu             sync.RWMutex
	roomRepository *RoomRepository
	placesClient   *services.PlacesClient
	rooms          map[string]*RoomConnections
	max            int
}

func NewRoomManager(max int, roomRepository *RoomRepository, placesClient *services.PlacesClient) *RoomManager {
	return &RoomManager{
		roomRepository: roomRepository,
		placesClient:   placesClient,
		rooms:          make(map[string]*RoomConnections),
		max:            max,
	}
}

func (rm *RoomManager) HandleVote(ctx context.Context, vote util.Vote) error {
	if vote.Room == "" {
		return fmt.Errorf("room is required")
	}
	if vote.Id == "" {
		return fmt.Errorf("vote id is required")
	}

	// only count ACCEPT votes
	if strings.ToUpper(vote.Result) != "ACCEPT" {
		return nil
	}

	// get client count from redis
	numClients, err := rm.roomRepository.GetRoomClientCount(ctx, vote.Room)
	if err != nil {
		return err
	}

	if numClients == 0 {
		return fmt.Errorf("room %s has no connected clients", vote.Room)
	}

	majorityFound, err := rm.roomRepository.IncrementAcceptVote(ctx, vote.Room, vote.Id, int(numClients))
	if err != nil {
		return err
	}

	if majorityFound {
		rm.roomRepository.PublishMajorityFound(ctx, vote.Room, vote.Id)
	}

	return nil
}

func (rm *RoomManager) StartEventListener() {
	ch := rm.roomRepository.SubscribeToRoomEvents()

	for msg := range ch {
		var event RoomEvent

		err := json.Unmarshal([]byte(msg.Payload), &event)
		if err != nil {
			log.Println("invalid event:", err)
			continue
		}

		rm.handleEvent(event)
	}
}

func (rm *RoomManager) handleEvent(event RoomEvent) {
	switch event.Type {
	case "room_started":
		fmt.Println("Received room_started event for room:", event.Room)
		rm.BroadcastRoomStarted(event.Room)
	case "majority_found":
		fmt.Println("Received majority_found event for room:", event.Room, "voteID:", event.VoteID)
		rm.BroadcastMajorityFound(event.Room, event.VoteID)
	}
}

func (rm *RoomManager) BroadcastRoomStarted(code string) {
	rm.mu.RLock()
	roomConnections, exists := rm.rooms[code]
	rm.mu.RUnlock()

	if !exists {
		fmt.Println("No active connections for room:", code)
		return
	}

	roomConnections.BroadcastRoomStarted()
}

func (rm *RoomManager) BroadcastMajorityFound(code string, voteID string) {
	rm.mu.RLock()
	roomConnections, exists := rm.rooms[code]
	rm.mu.RUnlock()

	if !exists {
		fmt.Println("No active connections for room:", code)
		return
	}

	roomConnections.BroadcastMajorityFound(voteID)
}

func (rm *RoomManager) GenerateRoomCode() string {
	return fmt.Sprintf("%06d", rand.Intn(rm.max))
}

func (rm *RoomManager) AddRoom() (string, error) {
	for {
		code := rm.GenerateRoomCode()

		roomCreated, err := rm.roomRepository.CreateRoom(code)
		if err != nil {
			return "", err
		}

		if roomCreated {
			return code, nil
		}

		// else: collision → try again
	}
}

func (rm *RoomManager) StartRoom(code string, filters util.Filters) (bool, error) {
	ctx := context.Background()

	cards, nextPageToken, err := rm.placesClient.FetchCards(filters, "")
	if err != nil {
		return false, fmt.Errorf("failed to fetch places: %w", err)
	}

	if len(cards) == 0 {
		return false, fmt.Errorf("no places found within the specified area")
	}

	if err := rm.roomRepository.SetRoomCards(ctx, code, cards); err != nil {
		return false, fmt.Errorf("failed to save room cards: %w", err)
	}

	if err := rm.roomRepository.SetPageToken(ctx, code, nextPageToken); err != nil {
		return false, fmt.Errorf("failed to save page token: %w", err)
	}

	err = rm.roomRepository.StartRoom(code)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (rm *RoomManager) ValidateRoomJoin(code string) (bool, error) {
	started, err := rm.roomRepository.CheckIfRoomStarted(code)
	if err != nil {
		return false, err
	}
	if started {
		return false, nil
	}

	return true, nil
}

func (rm *RoomManager) RegisterConn(code string, conn *websocket.Conn) error {
	exists, err := rm.roomRepository.CheckIfRoomExists(code)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("room not found")
	}

	rm.mu.Lock()
	roomConnections, exists := rm.rooms[code]
	if !exists {
		roomConnections = NewRoomConnections()
		rm.rooms[code] = roomConnections
	}
	rm.mu.Unlock()

	roomConnections.Add(conn)

	_, err = rm.roomRepository.IncrementRoomClientCount(context.Background(), code)
	return err
}

func (rm *RoomManager) UnregisterConn(code string, conn *websocket.Conn) {
	rm.mu.RLock()
	roomConnections, exists := rm.rooms[code]
	rm.mu.RUnlock()

	if !exists {
		return
	}

	roomConnections.Remove(conn)
	_, _ = rm.roomRepository.DecrementRoomClientCount(context.Background(), code)
}

func (rm *RoomManager) GetRoomCards(ctx context.Context, code string) ([]util.Card, error) {
	return rm.roomRepository.GetRoomCards(ctx, code)
}

func (rm *RoomManager) GetPhotoURL(photoName string) (string, error) {
	ctx := context.Background()

	// 1. Check if the photo URL has already been fetched and cached in the repository
	cachedURL, err := rm.roomRepository.GetCachedPhotoURL(ctx, photoName)
	if err == nil && cachedURL != "" {
		return cachedURL, nil
	}

	// 2. If not found in cache, call the placesClient
	photoURL, err := rm.placesClient.GetPhotoURL(photoName)
	if err != nil {
		return "", err
	}

	// 3. Save the newly fetched photo URL to the repository
	err = rm.roomRepository.SetCachedPhotoURL(ctx, photoName, photoURL)
	if err != nil {
		// Log the error, but do not fail the request since we successfully got the URL
		fmt.Printf("Warning: failed to cache photo URL for %s: %v\n", photoName, err)
	}

	return photoURL, nil
}
