package handlers

import (
	"backend/util"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// ==========================================
// Request DTO Structs (Consistent Naming)
// ==========================================

type StartRoomRequest struct {
	RoomCode string       `json:"roomCode"`
	Filters  util.Filters `json:"filters"`
}

type JoinRoomRequest struct {
	RoomCode string `json:"roomCode"`
}

// ==========================================
// HTTP Handlers
// ==========================================

func (s *Server) GetRoomCode(w http.ResponseWriter, req *http.Request) {
	code, err := s.RoomManager.AddRoom()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	m := util.Message{
		Header: "Room Code",
		Body:   &code,
	}

	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) HandleRoomJoin(w http.ResponseWriter, req *http.Request) {
	// 1. Parse using the new robust struct parser
	payload, err := parseJoinRoomRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Received request to join room:", payload.RoomCode)
	w.Header().Set("Content-Type", "application/json")

	joined, err := s.RoomManager.ValidateRoomJoin(payload.RoomCode)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(util.ErrorResponse{
			Header: "Join Room Error",
			Body:   err.Error(),
		})
		return
	}

	if !joined {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(util.ErrorResponse{
			Header: "Join Room Error",
			Body:   "room already started",
		})
		return
	}

	strJoined := strconv.FormatBool(joined)
	m := util.Message{
		Header: "Join Status",
		Body:   &strJoined,
	}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) HandleRoomStart(w http.ResponseWriter, req *http.Request) {
	payload, err := parseStartRoomRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("Attempting to start room: %s with filters: %+v\n", payload.RoomCode, payload.Filters)
	w.Header().Set("Content-Type", "application/json")

	started, err := s.RoomManager.StartRoom(payload.RoomCode, payload.Filters)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(util.ErrorResponse{
			Header: "Start Room Error",
			Body:   err.Error(),
		})
		return
	}

	strStarted := strconv.FormatBool(started)
	m := util.Message{
		Header: "Start Status",
		Body:   &strStarted,
	}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) HandleGetCardData(w http.ResponseWriter, req *http.Request) {
	// 1. Extract the room code from the URL query parameters (e.g., /get-card-data?code=123456)
	code := req.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing room code", http.StatusBadRequest)
		return
	}

	// 2. Fetch the cards from Redis using the request context
	// Note: If your RoomRepo is unexported inside RoomManager, you will need to add a
	// simple passthrough method called GetRoomCards on your RoomManager.
	cards, err := s.RoomManager.GetRoomCards(req.Context(), code)
	if err != nil {
		http.Error(w, "Failed to fetch cards: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Ensure we return an empty array `[]` instead of `null` if the room has no cards yet
	if cards == nil {
		cards = []util.Card{}
	}

	m := util.Message{
		Header: "CARD_DATA",
		Cards:  cards,
	}

	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ==========================================
// Request Parser Functions (Consistent Naming)
// ==========================================

func parseJoinRoomRequest(req *http.Request) (*JoinRoomRequest, error) {
	defer req.Body.Close()

	var payload JoinRoomRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		fmt.Printf("Error decoding join room payload: %v\n", err)
		return nil, err
	}

	if payload.RoomCode == "" {
		return nil, errors.New("empty room code")
	}

	return &payload, nil
}

func parseStartRoomRequest(req *http.Request) (*StartRoomRequest, error) {
	defer req.Body.Close()

	var payload StartRoomRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		fmt.Printf("Error decoding start room payload: %v\n", err)
		return nil, err
	}

	if payload.RoomCode == "" {
		return nil, errors.New("empty room code")
	}

	return &payload, nil
}
