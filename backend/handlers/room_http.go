package handlers

import (
	"backend/util"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"io"
)

func (s *Server) GetRoomCode(w http.ResponseWriter, req *http.Request) {
	// Inform client that the response type is JSON
	w.Header().Set("Content-Type", "application/json")
    // Set the HTTP status code (optional, http.StatusOK is 200).
	w.WriteHeader(http.StatusOK)
	var code = s.RoomService.GenerateCode()
	m := util.Message{
		Header: "Room Code", 
		Body: &code,
	}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) HandleRoomJoin(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	roomCode, err := parseRoomCode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		fmt.Print(roomCode)
	}

	// Inform client that the response type is JSON
	w.Header().Set("Content-Type", "application/json")

	joined, err := s.RoomService.JoinRoom(roomCode)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(util.ErrorResponse{
			Header: "Join Room Error",
			Body:   err.Error(),
		})
		return
	}

	if !joined {
		w.WriteHeader(http.StatusForbidden) // Room already started
		json.NewEncoder(w).Encode(util.ErrorResponse{
			Header: "Join Room Error",
			Body:   "room already started",
		})
		return
	}

	strJoined := strconv.FormatBool(joined)
	m := util.Message{
		Header: "Join Status", 
		Body: &strJoined,
	}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) HandleRoomStart(w http.ResponseWriter, req *http.Request) {
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	fmt.Printf("Raw request body: %s\n", string(bodyBytes))
	roomCode, err := parseRoomCode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Attempting to start room:", roomCode)
	w.Header().Set("Content-Type", "application/json")

	started, err := s.RoomService.StartRoom(roomCode)
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
		Body: &strStarted,
	}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func parseRoomCode(req *http.Request) (string, error) {
	defer req.Body.Close()
	fmt.Printf("Inside parseRoomCode...\n")
	//fmt.Printf("Parsing room code from request body...\n")
	//fmt.Printf("Request headers: %v\n", req.Header)
	//fmt.Printf("Raw request body: %v\n", req.Body)
	var roomCode string
	if err := json.NewDecoder(req.Body).Decode(&roomCode); err != nil {
		fmt.Printf("Error decoding room code: %v\n", err)
		return "", err
	}

	if roomCode == "" {
		return "", errors.New("empty room code")
	}
	fmt.Printf("Received room code: %s\n", roomCode)
	return roomCode, nil
}

func (s *Server) HandleGetCardData(w http.ResponseWriter, req *http.Request) {
	cards := []util.Card{
		{
			ID: "1",
			Title: "Sushi Express",
			Description: "Affordable conveyor-belt sushi chain popular for $2++ plates and a wide variety of Japanese dishes.",
		},
		{
			ID: "2",
			Title: "Eighteen Chefs",
			Description: "Casual Western-fusion restaurant famous for its 'Heart Attack Fried Rice' and hearty mains.",
		},
		{
			ID: "3",
			Title: "Seoul Garden",
			Description: "Korean BBQ buffet restaurant offering grill-it-yourself meats and hotpot options.",
		},
		{
			ID: "4",
			Title: "Ichiban Sushi",
			Description: "Family-friendly Japanese restaurant serving sushi, ramen, donburi and bento sets.",
		},
		{
			ID: "5",
			Title: "Swensen's",
			Description: "Classic Western restaurant known for fish & chips, burgers and ice cream desserts.",
		},
		{
			ID: "6",
			Title: "Pho Vietnam",
			Description: "Vietnamese restaurant serving pho noodle soups, banh mi and other traditional dishes.",
		},
		{
			ID: "7",
			Title: "Yakiniku Like",
			Description: "Japanese solo BBQ restaurant where diners grill individual meat sets quickly at their table.",
		},
		{
			ID: "8",
			Title: "Soup Restaurant",
			Description: "Singapore brand famous for its Samsui Ginger Chicken and traditional Chinese home-style dishes.",
		},
		{
			ID: "9",
			Title: "Kenny Rogers Roasters Express",
			Description: "Western chain known for roasted chicken, ribs, and hearty comfort-food sides.",
		},
		{
			ID: "10",
			Title: "Munchi Pancakes",
			Description: "Local snack stall selling traditional min jiang kueh pancakes with sweet fillings.",
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
	