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
		{
			ID:          "1",
			Title:       "Sushi Express",
			Category:    "Japanese Restaurant",
			PriceLevel:  "$",
			Rating:      4.2,
			ReviewCount: 1450,
			OpenNow:     true,
			Summary:     "Affordable conveyor-belt sushi chain popular for $2++ plates and a wide variety of Japanese dishes.",
			Address:     "10 Tampines Central 1, #04-12, Singapore 529536",
		},
		{
			ID:          "2",
			Title:       "Eighteen Chefs",
			Category:    "Western Fusion",
			PriceLevel:  "$$",
			Rating:      3.9,
			ReviewCount: 820,
			OpenNow:     true,
			Summary:     "Casual Western-fusion restaurant famous for its 'Heart Attack Fried Rice' and hearty mains.",
			Address:     "2 Handy Rd, #04-12 The Cathay, Singapore 229233",
		},
		{
			ID:          "3",
			Title:       "Seoul Garden",
			Category:    "Korean BBQ & Hotpot",
			PriceLevel:  "$$",
			Rating:      4.1,
			ReviewCount: 2100,
			OpenNow:     false,
			Summary:     "Korean BBQ buffet restaurant offering grill-it-yourself meats and hotpot options.",
			Address:     "200 Victoria St, #02-52 Bugis Junction, Singapore 188021",
		},
		{
			ID:          "4",
			Title:       "Ichiban Sushi",
			Category:    "Japanese Restaurant",
			PriceLevel:  "$$",
			Rating:      4.3,
			ReviewCount: 940,
			OpenNow:     true,
			Summary:     "Family-friendly Japanese restaurant serving sushi, ramen, donburi and bento sets.",
			Address:     "53 Ang Mo Kio Ave 3, #02-01 AMK Hub, Singapore 569933",
		},
		{
			ID:          "5",
			Title:       "Swensen's",
			Category:    "Western Restaurant",
			PriceLevel:  "$$",
			Rating:      4.0,
			ReviewCount: 1750,
			OpenNow:     true,
			Summary:     "Classic Western restaurant known for fish & chips, burgers and ice cream desserts.",
			Address:     "68 Orchard Rd, #03-23 Plaza Singapura, Singapore 238839",
		},
		{
			ID:          "6",
			Title:       "Pho Vietnam",
			Category:    "Vietnamese Restaurant",
			PriceLevel:  "$",
			Rating:      4.5,
			ReviewCount: 310,
			OpenNow:     true,
			Summary:     "Vietnamese restaurant serving pho noodle soups, banh mi and other traditional dishes.",
			Address:     "200 Turf Club Rd, Singapore 287994",
		},
		{
			ID:          "7",
			Title:       "Yakiniku Like",
			Category:    "Japanese Solo BBQ",
			PriceLevel:  "$$",
			Rating:      4.4,
			ReviewCount: 1120,
			OpenNow:     false,
			Summary:     "Japanese solo BBQ restaurant where diners grill individual meat sets quickly at their table.",
			Address:     "10 Paya Lebar Rd, #B1-28 Paya Lebar Quarter, Singapore 409057",
		},
		{
			ID:          "8",
			Title:       "Soup Restaurant",
			Category:    "Chinese Restaurant",
			PriceLevel:  "$$$",
			Rating:      4.2,
			ReviewCount: 680,
			OpenNow:     true,
			Summary:     "Singapore brand famous for its Samsui Ginger Chicken and traditional Chinese home-style dishes.",
			Address:     "1 HarbourFront Walk, #02-141 VivoCity, Singapore 098585",
		},
		{
			ID:          "9",
			Title:       "Kenny Rogers Roasters Express",
			Category:    "Western Halal Chain",
			PriceLevel:  "$$",
			Rating:      3.7,
			ReviewCount: 240,
			OpenNow:     true,
			Summary:     "Western chain known for roasted chicken, ribs, and hearty comfort-food sides.",
			Address:     "1 Pasir Ris Close, #02-336 Downtown East, Singapore 519599",
		},
		{
			ID:          "10",
			Title:       "Munchi Pancakes",
			Category:    "Local Snack Stall",
			PriceLevel:  "$",
			Rating:      4.6,
			ReviewCount: 530,
			OpenNow:     true,
			Summary:     "Local snack stall selling traditional min jiang kueh pancakes with sweet fillings.",
			Address:     "51 Yishun Ave 11, #01-43 Yishun Park Hawker Centre, Singapore 768867",
		},
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
