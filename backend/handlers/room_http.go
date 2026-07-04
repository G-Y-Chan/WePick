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
	cards := []util.Card{
		{ID: "1", Title: "Sushi Express", Description: "Affordable conveyor-belt sushi chain popular for $2++ plates and a wide variety of Japanese dishes."},
		{ID: "2", Title: "Eighteen Chefs", Description: "Casual Western-fusion restaurant famous for its 'Heart Attack Fried Rice' and hearty mains."},
		{ID: "3", Title: "Seoul Garden", Description: "Korean BBQ buffet restaurant offering grill-it-yourself meats and hotpot options."},
		{ID: "4", Title: "Ichiban Sushi", Description: "Family-friendly Japanese restaurant serving sushi, ramen, donburi and bento sets."},
		{ID: "5", Title: "Swensen's", Description: "Classic Western restaurant known for fish & chips, burgers and ice cream desserts."},
		{ID: "6", Title: "Pho Vietnam", Description: "Vietnamese restaurant serving pho noodle soups, banh mi and other traditional dishes."},
		{ID: "7", Title: "Yakiniku Like", Description: "Japanese solo BBQ restaurant where diners grill individual meat sets quickly at their table."},
		{ID: "8", Title: "Soup Restaurant", Description: "Singapore brand famous for its Samsui Ginger Chicken and traditional Chinese home-style dishes."},
		{ID: "9", Title: "Kenny Rogers Roasters Express", Description: "Western chain known for roasted chicken, ribs, and hearty comfort-food sides."},
		{ID: "10", Title: "Munchi Pancakes", Description: "Local snack stall selling traditional min jiang kueh pancakes with sweet fillings."},
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
